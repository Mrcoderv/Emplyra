package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/emplyra/backend/internal/auditmanager"
	"github.com/emplyra/backend/internal/config"
	"github.com/emplyra/backend/internal/models"
	"github.com/emplyra/backend/internal/repositories"
	"github.com/emplyra/backend/internal/utils"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrFileTooLarge     = errors.New("file too large")
)

type DocumentService struct {
	docs  *repositories.DocumentRepository
	emp   *repositories.EmployeeRepository
	cfg   *config.Config
	audit *auditmanager.Service
}

func NewDocumentService(docs *repositories.DocumentRepository, emp *repositories.EmployeeRepository, cfg *config.Config, audit *auditmanager.Service) *DocumentService {
	return &DocumentService{docs: docs, emp: emp, cfg: cfg, audit: audit}
}

func (s *DocumentService) Upload(tenantID, employeeID, title, docType string, uploadedBy *string, fh *multipart.FileHeader) (*models.Document, error) {
	if _, err := s.emp.FindByID(tenantID, employeeID); err != nil {
		return nil, ErrEmployeeNotFound
	}
	if fh == nil {
		return nil, errors.New("no file provided")
	}
	if fh.Size > s.cfg.MaxUploadSizeMB*1024*1024 {
		return nil, ErrFileTooLarge
	}
	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dir := filepath.Join(s.cfg.UploadDir, employeeID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	ext := filepath.Ext(fh.Filename)
	storeName := randomHex(16) + ext
	path := filepath.Join(dir, storeName)
	dst, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(path)
		return nil, err
	}
	if err := dst.Close(); err != nil {
		os.Remove(path)
		return nil, err
	}

	doc := &models.Document{
		TenantID:   tenantID,
		EmployeeID: employeeID,
		Title:      title,
		Type:       models.DocumentType(orStr(docType, string(models.DocOther))),
		FilePath:   path,
		MimeType:   fh.Header.Get("Content-Type"),
		SizeBytes:  fh.Size,
		Status:     models.DocumentActive,
		UploadedBy: uploadedBy,
	}
	if err := s.docs.Create(doc); err != nil {
		os.Remove(path)
		return nil, err
	}
	if uploadedBy != nil {
		s.audit.Record(*uploadedBy, models.ActionCreate, "document", doc.ID, "", "", map[string]string{"title": title})
	}
	return doc, nil
}

func (s *DocumentService) Get(tenantID, id string) (*models.Document, error) {
	d, err := s.docs.FindByID(tenantID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return d, nil
}

func (s *DocumentService) List(tenantID string, p utils.Pagination, employeeID, docType, status string) ([]models.Document, int64, error) {
	return s.docs.List(tenantID, p, employeeID, docType, status)
}

func (s *DocumentService) Delete(tenantID, id string, actorID *string) error {
	d, err := s.docs.FindByID(tenantID, id)
	if err != nil {
		return ErrNotFound
	}
	if err := s.docs.Delete(tenantID, id); err != nil {
		return err
	}
	if d.FilePath != "" {
		_ = os.Remove(d.FilePath)
	}
	if actorID != nil {
		s.audit.Record(*actorID, models.ActionDelete, "document", id, "", "", nil)
	}
	return nil
}

func (s *DocumentService) DownloadPath(tenantID, id string) (string, string, error) {
	d, err := s.docs.FindByID(tenantID, id)
	if err != nil {
		return "", "", ErrNotFound
	}
	return d.FilePath, d.MimeType, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return hex.EncodeToString(b)
}
