package repository

import (
	"context"
	"fmt"

	"iclassroom/backend/internal/domain"
)

// CreateImages inserts submission image metadata in a single transaction.
func (r *SubmissionRepository) CreateImages(ctx context.Context, submissionID int64, images []domain.SubmissionImage) ([]domain.SubmissionImage, error) {
	if len(images) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: begin image tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const insertImage = `INSERT INTO submission_images
(submission_id, file_url, file_path, file_name, file_size, mime_type)
VALUES (?, ?, ?, ?, ?, ?)`

	for i := range images {
		res, err := tx.ExecContext(
			ctx,
			insertImage,
			submissionID,
			images[i].FileURL,
			images[i].FilePath,
			images[i].FileName,
			images[i].FileSize,
			images[i].MimeType,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: insert submission image: %w", err)
		}

		imageID, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("repository: submission image last insert id: %w", err)
		}
		images[i].ID = imageID
		images[i].SubmissionID = submissionID
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit image tx: %w", err)
	}

	return images, nil
}

// DeleteByID removes one submission row. Linked images are removed by cascade.
func (r *SubmissionRepository) DeleteByID(ctx context.Context, submissionID int64) error {
	const q = `DELETE FROM submissions WHERE id = ?`

	res, err := r.db.ExecContext(ctx, q, submissionID)
	if err != nil {
		return fmt.Errorf("repository: delete submission: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: delete submission rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}
