package database

import (
	"database/sql"
	"fmt"
	"github.com/buts00/Graph/internal/app/graph"
	"github.com/buts00/Graph/internal/config"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

const (
	edgesTable     = "edges"
	selectAllEdges = "SELECT * FROM " + edgesTable + ";"
	deleteAllEdges = "DELETE FROM " + edgesTable + ";"
	insertEdge     = "INSERT INTO " + edgesTable + " (source_node_id, destination_node_id, weight) VALUES ($1, $2, $3) RETURNING edge_id;"
	deleteEdge     = "DELETE FROM " + edgesTable + " WHERE source_node_id = $1 AND destination_node_id = $2  AND weight = $3 RETURNING edge_id;"
	countEdges     = "SELECT COUNT(*) FROM " + edgesTable + " WHERE source_node_id = $1 AND destination_node_id = $2 AND weight = $3;"
)

func NewPostgresDB(config config.Config) (*PostgresDB, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DbName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{db}, nil
}

func Edges(db *PostgresDB) (graph.Graph, error) {
	row, err := db.DB.Query(selectAllEdges)
	if err != nil {
		return graph.Graph{}, err
	}

	var cur graph.Graph
	for row.Next() {
		var id, source, destination, weight int
		if err := row.Scan(&id, &source, &destination, &weight); err != nil {
			return graph.Graph{}, err
		}
		cur.Edges = append(cur.Edges, graph.Edge{Id: id, Source: &source, Destination: &destination, Weight: weight})
	}

	return cur, nil
}

func AddEdge(db *PostgresDB, edge graph.Edge) (int, error) {

	var id int
	err := db.DB.QueryRow(insertEdge, edge.Source, edge.Destination, edge.Weight).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func DeleteEdge(db *PostgresDB, edge graph.Edge) (int, error) {

	var id int
	err := db.DB.QueryRow(deleteEdge, edge.Source, edge.Destination, edge.Weight).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func IsEdgeExist(db *PostgresDB, edge graph.Edge) (bool, bool, error) {
	var (
		count, reversedCount int
	)

	err := db.DB.QueryRow(countEdges, edge.Source, edge.Destination, edge.Weight).Scan(&count)
	if err != nil {
		return false, false, err
	}

	err = db.DB.QueryRow(countEdges, edge.Destination, edge.Source, edge.Weight).Scan(&reversedCount)

	if err != nil {
		return false, false, err
	}

	return count > 0, reversedCount > 0, nil
}

func ClearGraph(db *PostgresDB) error {
	_, err := db.DB.Exec(deleteAllEdges)
	return err
}
