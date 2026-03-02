#!/bin/bash
set -e


echo "NEO4J_AUTH: $NEO4J_AUTH"
echo "NEO4J_AUTH_FILE: $NEO4J_AUTH_FILE"      

AUTH_STR="$NEO4J_AUTH"

USER=$(echo "$AUTH_STR" | cut -d'/' -f1)
PASS=$(echo "$AUTH_STR" | cut -d'/' -f2)

echo "Connecting to Neo4j at neo4j:7687 as user: $USER/$PASS"

cypher-shell -a neo4j://neo4j:7687 -u "$USER" -p "$PASS" \
"CREATE CONSTRAINT n10s_unique_uri IF NOT EXISTS FOR (r:Resource) REQUIRE r.uri IS UNIQUE;"

CONFIG_EXISTS=$(cypher-shell -a neo4j://neo4j:7687 -u "$USER" -p "$PASS" 'CALL n10s.graphconfig.show()' | grep -c 'handleVocabUris' || true)

if [ "$CONFIG_EXISTS" -eq 0 ]; then
    echo "Graph is empty, running n10s.graphconfig.init()..."
    cypher-shell -a neo4j://neo4j:7687 -u "$USER" -p "$PASS" \
    "CALL n10s.graphconfig.init({handleVocabUris: 'MAP', applyNeo4jNaming: true});"
else
    echo "n10s is already prepared, skipping."
fi

echo "n10s preparing finished!"