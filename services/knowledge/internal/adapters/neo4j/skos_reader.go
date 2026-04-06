package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// GetTrees извлекает все фасеты и строит из них деревья
func (c *Client) GetTrees(ctx context.Context) ([]core.SKOSFacet, error) {
	session := c.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Запрос собирает фасеты, концепты, их родителей и связанные концепты
		query := `
			MATCH (s:ConceptScheme)
			OPTIONAL MATCH (c:Concept)-[:IN_SCHEME]->(s)
			OPTIONAL MATCH (c)-[:BROADER]->(p:Concept)
			OPTIONAL MATCH (c)-[:RELATED]-(r:Concept)
			RETURN 
				s.uri AS schemeID, s.prefLabel AS schemeLabel,
				c.uri AS conceptID, c.prefLabel AS conceptLabel,
				p.uri AS parentID,
				collect(r.uri) AS relatedIDs
		`
		res, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		facetsMap := make(map[string]*core.SKOSFacet)
		nodesMap := make(map[string]*core.SKOSTreeNode)
		parentMap := make(map[string]string) // childID -> parentID
		schemeMap := make(map[string]string) // conceptID -> schemeID

		for res.Next(ctx) {
			rec := res.Record()
			schemeID, _ := rec.Get("schemeID")
			schemeLabel, _ := rec.Get("schemeLabel")
			
			sID := toString(schemeID)
			
			if _, ok := facetsMap[sID]; !ok {
				facetsMap[sID] = &core.SKOSFacet{
					ID:       sID,
					Label:    toString(schemeLabel),
					Concepts: make([]*core.SKOSTreeNode, 0),
				}
			}

			cIDRaw, _ := rec.Get("conceptID")
			if cIDRaw == nil {
				continue // Пустой фасет
			}
			
			cID := toString(cIDRaw)
			cLabel, _ := rec.Get("conceptLabel")
			pIDRaw, _ := rec.Get("parentID")
			relRaw, _ := rec.Get("relatedIDs")

			var related []string
			if relList, ok := relRaw.([]any); ok {
				for _, r := range relList {
					related = append(related, toString(r))
				}
			}

			if _, ok := nodesMap[cID]; !ok {
				nodesMap[cID] = &core.SKOSTreeNode{
					ID:      cID,
					Label:   toString(cLabel),
					Related: related,
				}
				schemeMap[cID] = sID
				if pIDRaw != nil {
					parentMap[cID] = toString(pIDRaw)
				}
			}
		}

		// Сборка деревьев
		for cID, node := range nodesMap {
			if parentID, hasParent := parentMap[cID]; hasParent {
				if parentNode, ok := nodesMap[parentID]; ok {
					parentNode.Children = append(parentNode.Children, node)
				}
			} else {
				// Если нет родителя, это корень фасета
				sID := schemeMap[cID]
				facetsMap[sID].Concepts = append(facetsMap[sID].Concepts, node)
			}
		}

		var facets []core.SKOSFacet
		for _, f := range facetsMap {
			facets = append(facets, *f)
		}

		return facets, nil
	})

	if err != nil {
		return nil, fmt.Errorf("execute get trees query: %w", err)
	}

	return result.([]core.SKOSFacet), nil
}