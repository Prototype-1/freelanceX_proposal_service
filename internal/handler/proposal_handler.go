package handler

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/model"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/service"
	pb "github.com/Prototype-1/freelanceX_proposal_service/proto"
	"github.com/Prototype-1/freelanceX_proposal_service/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
	"time"
	"log"
	"google.golang.org/grpc/metadata"
	"strings"
)

type ProposalHandler struct {
	pb.UnimplementedProposalServiceServer
	service *service.ProposalService
}

func NewProposalHandler(service *service.ProposalService) *ProposalHandler {
	return &ProposalHandler{service: service}
}

func extractRole(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	roles := md.Get("role")
	if len(roles) == 0 {
		return ""
	}
	return strings.ToLower(roles[0])
}

func (h *ProposalHandler) CreateProposal(ctx context.Context, req *pb.CreateProposalRequest) (*pb.CreateProposalResponse, error) {
	if extractRole(ctx) != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "only freelancers can create proposals")
	}

	var deadline time.Time
	var err error

	if req.GetDeadlineStr() != "" {
		deadline, err = time.Parse(time.RFC3339, req.GetDeadlineStr())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid deadline format: %v", err)
		}
	} else if req.GetDeadline() != nil {
		deadline = req.GetDeadline().AsTime()
	}else {
    deadline = time.Now()
}

var sections []model.Section

	if req.GetTemplateId() != "" {
		templateID, err := primitive.ObjectIDFromHex(req.GetTemplateId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid template ID: %v", err)
		}
		template, err := h.service.GetTemplateByID(ctx, templateID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to fetch template: %v", err)
		}
		sections = template.Sections
	} 

	proposal := model.Proposal{
		ClientID:     req.GetClientId(),
		FreelancerID: req.GetFreelancerId(),
		Title:        req.GetTitle(),
		Content:      req.GetContent(),
		Status: "draft",
		Version:      1,
		Deadline:     deadline,
		Sections:     sections,
	}

	createdProposal, err := h.service.CreateProposal(ctx, proposal)
	if err != nil {
		return nil, err
	}

	event := kafka.ProposalEvent{
		ProposalID:   createdProposal.ID.Hex(),
		ClientID:     createdProposal.ClientID,
		FreelancerID: createdProposal.FreelancerID,
		Title:        createdProposal.Title,
		EventType:    "proposal.created",
		Status:       "sent",
	}

	go func() {
		err := kafka.ProduceProposalEvent("localhost:9092", "proposal-events", event)
		if err != nil {
			log.Printf("failed to produce proposal.created event: %v", err)
		}
	}()

	return &pb.CreateProposalResponse{
		ProposalId: createdProposal.ID.Hex(),
		Status:     "created",
	}, nil
}

func (h *ProposalHandler) GetProposalByID(ctx context.Context, req *pb.GetProposalRequest) (*pb.GetProposalResponse, error) {
	role := extractRole(ctx)
	if role != "freelancer" && role != "client" {
		return nil, status.Error(codes.PermissionDenied, "you are unauthorized to get proposal")
	}

	proposal, err := h.service.GetProposalByID(ctx, req.GetProposalId())
	if err != nil {
		return nil, err
	}

	return &pb.GetProposalResponse{
		ProposalId:    proposal.ID.Hex(),
		ClientId:      proposal.ClientID,
		FreelancerId:  proposal.FreelancerID,
		TemplateId:    proposal.TemplateID.Hex(),
		Title:         proposal.Title,
		Content:       proposal.Content,
		Status:        proposal.Status,
		Version:       int32(proposal.Version),
		Deadline: timestamppb.New(proposal.Deadline),
		DeadlineStr:  proposal.Deadline.Format(time.RFC3339),
		CreatedAt:     timestamppb.New(proposal.CreatedAt),
		UpdatedAt:     timestamppb.New(proposal.UpdatedAt),
	}, nil
}

func (h *ProposalHandler) UpdateProposal(ctx context.Context, req *pb.UpdateProposalRequest) (*pb.UpdateProposalResponse, error) {
	
	role := extractRole(ctx)
	update := model.Proposal{
		Title:   req.GetTitle(),
		Content: req.GetContent(),
	}

	if req.GetDeadline() != nil {
		update.Deadline = req.GetDeadline().AsTime()
	}

	if role == "client" {
		if req.GetTitle() != "" || req.GetContent() != "" || req.GetDeadline() != nil {
			return nil, status.Error(codes.PermissionDenied, "clients can only update status")
		}
newStatus := req.GetStatus()
if newStatus != "accepted" && newStatus != "rejected" {
    return nil, status.Error(codes.PermissionDenied, "clients can only set status to accepted or rejected")
}
update.Status = newStatus
		
	}

	if role != "freelancer" && role != "client" {
		return nil, status.Error(codes.PermissionDenied, "unauthorized to update proposal")
	}

	updatedProposal, err := h.service.UpdateProposal(ctx, req.GetProposalId(), update)
	if err != nil {
		return nil, err
	}

	if update.Status != "" {
		event := kafka.ProposalEvent{
			ProposalID:   updatedProposal.ID.Hex(),
			ClientID:     updatedProposal.ClientID,
			FreelancerID: updatedProposal.FreelancerID,
			Title:        updatedProposal.Title,
			EventType:    "proposal.updated",
			Status:       updatedProposal.Status,
		}

		go func() {
			if err := kafka.ProduceProposalEvent("localhost:9092", "proposal-events", event); err != nil {
				log.Printf("failed to produce proposal.updated event: %v", err)
			}
		}()
	}

	return &pb.UpdateProposalResponse{
		ProposalId: updatedProposal.ID.Hex(),
		Status:     "updated",
		NewVersion: int32(updatedProposal.Version),
	}, nil
}

func (h *ProposalHandler) SaveTemplate(ctx context.Context, req *pb.SaveTemplateRequest) (*pb.SaveTemplateResponse, error) {
	if extractRole(ctx) != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "only freelancers can save templates")
	}

	template := model.Template{
		OwnerID: req.GetFreelancerId(),
		Title:   req.GetTitle(),
		Sections: []model.Section{
			{
				Heading: "Default Heading",
				Body:    req.GetContent(),
			},
		},
	}

	_, err := h.service.SaveTemplate(ctx, template)
	if err != nil {
		return nil, err
	}

	return &pb.SaveTemplateResponse{
		Status: "created",
	}, nil
}

func (h *ProposalHandler) GetTemplatesForFreelancer(ctx context.Context, req *pb.GetTemplatesRequest) (*pb.GetTemplatesResponse, error) {
	if extractRole(ctx) != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "only freelancers can view templates")
	}

	templates, err := h.service.GetTemplatesForFreelancer(ctx, req.GetFreelancerId())
	if err != nil {
		return nil, err
	}

	var pbTemplates []*pb.Template
	for _, template := range templates {
		var sectionsContent string
		for _, section := range template.Sections {
			sectionsContent += section.Heading + ": " + section.Body + "\n"
		}

		pbTemplates = append(pbTemplates, &pb.Template{
			TemplateId: template.ID.Hex(),
			Title:      template.Title,
			Content:    sectionsContent, 
		})
	}

	return &pb.GetTemplatesResponse{
		Templates: pbTemplates,
	}, nil
}

func (h *ProposalHandler) ListProposals(ctx context.Context, req *pb.ListProposalsRequest) (*pb.ListProposalsResponse, error) {
	if extractRole(ctx) != "admin" {
		return nil, status.Error(codes.PermissionDenied, "only admins can list proposals")
	}
	
    filters := make(map[string]interface{})
    if req.GetClientId() != "" {
        filters["client_id"] = req.GetClientId()
    }
    if req.GetFreelancerId() != "" {
        filters["freelancer_id"] = req.GetFreelancerId()
    }
    if req.GetStatus() != "" {
        filters["status"] = req.GetStatus()
    }

    proposals, err := h.service.GetProposals(ctx, filters, req.GetSkip(), req.GetLimit())
    if err != nil {
        return nil, err
    }

    var protoProposals []*pb.Proposal
    for _, p := range proposals {
        protoProposals = append(protoProposals, &pb.Proposal{
            ProposalId:    p.ID.Hex(),
            ClientId:      p.ClientID,
            FreelancerId:  p.FreelancerID,
            TemplateId:    p.TemplateID.Hex(),
            Title:         p.Title,
            Content:       p.Content,
            Status:        p.Status,
            Version:       int32(p.Version),
            CreatedAt:     timestamppb.New(p.CreatedAt),
            UpdatedAt:     timestamppb.New(p.UpdatedAt),
        })
    }

    return &pb.ListProposalsResponse{
        Proposals: protoProposals,
    }, nil
}



