package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/domain"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type nodeRepository struct {
	db *sql.DB
}

func NewNodeRepository(db *sql.DB) domain.NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) Create(ctx context.Context, node *entity.Node) error {
	query := `
		INSERT INTO nodes (id, name, description, type, priority_startup, fields, imports, exports, server_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		node.ID,
		node.Name,
		node.Description,
		node.Type,
		node.PriorityStartup,
		node.Fields,
		node.Imports,
		node.Exports,
		node.ServerID,
		node.CreatedAt,
		node.UpdatedAt,
	)
	return err
}

func (r *nodeRepository) Get(ctx context.Context, id string) (*entity.Node, error) {
	query := `SELECT id, name, description, type, priority_startup, fields, imports, exports, server_id, created_at, updated_at FROM nodes WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var node entity.Node
	var typeStr string
	err := row.Scan(
		&node.ID,
		&node.Name,
		&node.Description,
		&typeStr,
		&node.PriorityStartup,
		&node.Fields,
		&node.Imports,
		&node.Exports,
		&node.ServerID,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	node.Type = entity.NodeType(typeStr)
	return &node, nil
}

func (r *nodeRepository) List(ctx context.Context, offset, limit int) ([]*entity.Node, int64, error) {
	query := `SELECT id, name, description, type, priority_startup, fields, imports, exports, server_id, created_at, updated_at FROM nodes LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var nodes []*entity.Node
	for rows.Next() {
		var node entity.Node
		var typeStr string
		err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Description,
			&typeStr,
			&node.PriorityStartup,
			&node.Fields,
			&node.Imports,
			&node.Exports,
			&node.ServerID,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		node.Type = entity.NodeType(typeStr)
		nodes = append(nodes, &node)
	}

	var count int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM nodes`).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return nodes, count, nil
}

func (r *nodeRepository) Update(ctx context.Context, node *entity.Node) error {
	query := `
		UPDATE nodes SET name = ?, description = ?, type = ?, priority_startup = ?, fields = ?, imports = ?, exports = ?, server_id = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		node.Name,
		node.Description,
		node.Type,
		node.PriorityStartup,
		node.Fields,
		node.Imports,
		node.Exports,
		node.ServerID,
		time.Now(),
		node.ID,
	)
	return err
}

func (r *nodeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM nodes WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *nodeRepository) GetByServerID(ctx context.Context, serverID string) ([]*entity.Node, error) {
	query := `SELECT id, name, description, type, priority_startup, fields, imports, exports, server_id, created_at, updated_at FROM nodes WHERE server_id = ?`
	rows, err := r.db.QueryContext(ctx, query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*entity.Node
	for rows.Next() {
		var node entity.Node
		var typeStr string
		err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Description,
			&typeStr,
			&node.PriorityStartup,
			&node.Fields,
			&node.Imports,
			&node.Exports,
			&node.ServerID,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		node.Type = entity.NodeType(typeStr)
		nodes = append(nodes, &node)
	}
	return nodes, nil
}
