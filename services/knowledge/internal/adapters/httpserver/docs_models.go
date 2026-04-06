// Package httpserver содержит адаптер для HTTP сервера, который обрабатывает запросы и возвращает ответы в формате JSON.
package httpserver

import "github.com/penkovgd/erudition-app/services/knowledge/internal/core"

// ErrorResponse описывает структуру ответа при ошибке.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SKOSResponse описывает структуру ответа для запроса SKOS деревьев.
type SKOSResponse struct {
	// Здесь мы предполагаем, что core.Facet (или как называется ваше дерево) корректно сериализуется
	Facets []Facet `json:"facets"`
}

// Facet и Concept описывают структуру данных для SKOS деревьев, которые мы возвращаем в ответе.
type Facet struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Concepts []Concept `json:"concepts"`
}

// Concept описывает структуру данных для концептов внутри SKOS деревьев.
type Concept struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Related  []string  `json:"related,omitempty"`
	Children []Concept `json:"children,omitempty"`
}

// TopicsResponse описывает структуру ответа для запроса топиков с фильтрами.
type TopicsResponse struct {
	Topics []core.Topic `json:"topics"` // Если core.Topic содержит json теги, swag их прочитает
}

// SyncResponse описывает структуру ответа для запроса синхронизации данных.
type SyncResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Synchronization completed successfully"`
}
