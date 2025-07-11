package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/status"
    "google.golang.org/grpc/codes"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/repository"
)

type ProposalService struct {
	repo *repository.ProposalRepository
}

func NewProposalService(repo *repository.ProposalRepository) *ProposalService {
	return &ProposalService{repo: repo}
}

func (s *ProposalService) CreateProposal(ctx context.Context, proposal model.Proposal) (*model.Proposal, error) {
	log.Printf("Creating proposal with title: %s, content: %s, sections: %+v", proposal.Title, proposal.Content, proposal.Sections)

	if proposal.ClientID == "" || proposal.FreelancerID == "" || proposal.Title == "" {
		return nil, errors.New("missing required fields")
	}
	return s.repo.CreateProposal(ctx, proposal)
}

func (s *ProposalService) GetProposalByID(ctx context.Context, id string) (*model.Proposal, error) {
	return s.repo.GetProposalByID(ctx, id)
}

func (s *ProposalService) UpdateProposal(ctx context.Context, id string, updatedProposal model.Proposal) (*model.Proposal, error) {

	if !updatedProposal.Deadline.IsZero() && updatedProposal.Deadline.Before(time.Now()) {
		return nil, fmt.Errorf("cannot set the deadline to a past date")
	}
	validStatuses := map[string]bool{
		"draft":    true,
		"sent":     true,
		"accepted": true,
		"rejected": true,
		"expired": true,
	}
	
	if updatedProposal.Status != "" && !validStatuses[updatedProposal.Status] {
		return nil, fmt.Errorf("invalid status: %s", updatedProposal.Status)
	}

	updatedProposal.UpdatedAt = time.Now()
	return s.repo.UpdateProposal(ctx, id, updatedProposal)
}

func (s *ProposalService) SaveTemplate(ctx context.Context, template model.Template) (*model.Template, error) {
	if template.OwnerID == "" || template.Title == "" {
		return nil, errors.New("missing required fields for template")
	}
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	return s.repo.SaveTemplate(ctx, template)
}

func (s *ProposalService) GetTemplatesForFreelancer(ctx context.Context, freelancerID string) ([]*model.Template, error) {
	templates, err := s.repo.GetTemplatesForFreelancer(ctx, freelancerID)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (s *ProposalService) GetTemplateByID(ctx context.Context, id primitive.ObjectID) (*model.Template, error) {
	template, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "template not found")
		}
		return nil, err
	}
	return template, nil
}

func (s *ProposalService) GetProposals(ctx context.Context, filters map[string]interface{}, skip, limit int64) ([]*model.Proposal, error) {
	if filters == nil {
		filters = make(map[string]interface{})
	}
	proposals, err := s.repo.GetProposals(ctx, filters, skip, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve proposals: %w", err)
	}
	return proposals, nil
}
