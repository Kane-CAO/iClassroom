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
	const insertAttachment = `INSERT INTO submission_attachments
(submission_id, kind, file_url, file_path, original_file_name, stored_file_name, file_size, mime_type)
VALUES (?, 'image', ?, ?, ?, ?, ?, ?)`

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

		if _, err := tx.ExecContext(
			ctx,
			insertAttachment,
			submissionID,
			images[i].FileURL,
			images[i].FilePath,
			images[i].FileName,
			images[i].FileName,
			images[i].FileSize,
			images[i].MimeType,
		); err != nil {
			return nil, fmt.Errorf("repository: insert image attachment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit image tx: %w", err)
	}

	return images, nil
}

func (r *SubmissionRepository) CreateFiles(ctx context.Context, submissionID int64, files []domain.SubmissionFile) ([]domain.SubmissionFile, error) {
	if len(files) == 0 {
		return nil, nil
	}

	const q = `INSERT INTO submission_attachments
(submission_id, kind, file_url, file_path, original_file_name, stored_file_name, file_size, mime_type)
VALUES (?, 'file', ?, ?, ?, ?, ?, ?)`
	for i := range files {
		res, err := r.db.ExecContext(
			ctx,
			q,
			submissionID,
			files[i].FileURL,
			files[i].FilePath,
			files[i].OriginalFileName,
			files[i].StoredFileName,
			files[i].FileSize,
			files[i].MimeType,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: insert submission file: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("repository: submission file last insert id: %w", err)
		}
		files[i].ID = id
		files[i].SubmissionID = submissionID
		files[i].Kind = "file"
	}
	return files, nil
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
