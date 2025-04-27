package repository

import (
	"context"
	"fmt"
	"time"
	"github.com/Prototype-1/freelanceX_proposal_service/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProposalRepository struct {
	client *mongo.Client
}

func NewProposalRepository(client *mongo.Client) *ProposalRepository {
	return &ProposalRepository{client: client}
}

func (r *ProposalRepository) CreateProposal(ctx context.Context, proposal model.Proposal) (*model.Proposal, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("proposals")
	proposal.ID = primitive.NewObjectID()
	proposal.CreatedAt = time.Now()
	proposal.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	return &proposal, nil
}

func (r *ProposalRepository) GetProposalByID(ctx context.Context, proposalID string) (*model.Proposal, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("proposals")
	objID, err := primitive.ObjectIDFromHex(proposalID)
	if err != nil {
		return nil, fmt.Errorf("invalid proposal ID: %w", err)
	}

	var proposal model.Proposal
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&proposal)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("proposal with ID %s not found: %w", proposalID, err)
		}
		return nil, fmt.Errorf("failed to retrieve proposal: %w", err)
	}

	return &proposal, nil
}

func (r *ProposalRepository) UpdateProposal(ctx context.Context, proposalID string, update model.Proposal) (*model.Proposal, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("proposals")
	objID, err := primitive.ObjectIDFromHex(proposalID)
	if err != nil {
		return nil, fmt.Errorf("invalid proposal ID: %w", err)
	}

	updateFields := bson.M{
		"title":        update.Title,
		"content":      update.Content,
		"status":       update.Status,
		"updated_at":   time.Now(),
		"version":      bson.M{"$inc": 1},
	}

	updateResult := collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updateFields, "$inc": bson.M{"version": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updatedProposal model.Proposal
	if err := updateResult.Decode(&updatedProposal); err != nil {
		return nil, fmt.Errorf("failed to decode updated proposal: %w", err)
	}

	return &updatedProposal, nil
}

func (r *ProposalRepository) GetProposals(ctx context.Context, filters map[string]interface{}, skip, limit int64) ([]*model.Proposal, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("proposals")

	filter := bson.M{}
	for key, value := range filters {
		filter[key] = value
	}
	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query proposals: %w", err)
	}
	defer cursor.Close(ctx)

	var proposals []*model.Proposal
	for cursor.Next(ctx) {
		var proposal model.Proposal
		if err := cursor.Decode(&proposal); err != nil {
			return nil, fmt.Errorf("failed to decode proposal: %w", err)
		}
		proposals = append(proposals, &proposal)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return proposals, nil
}

func (r *ProposalRepository) SaveTemplate(ctx context.Context, template model.Template) (*model.Template, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("templates")
	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	return &template, nil
}

func (r *ProposalRepository) GetTemplatesForFreelancer(ctx context.Context, freelancerID string) ([]*model.Template, error) {
	collection := r.client.Database("freelanceX_proposals").Collection("templates")

	var templates []*model.Template
	cursor, err := collection.Find(ctx, bson.M{"owner_id": freelancerID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve templates for freelancer %s: %w", freelancerID, err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var template model.Template
		if err := cursor.Decode(&template); err != nil {
			return nil, fmt.Errorf("failed to decode template: %w", err)
		}
		templates = append(templates, &template)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %w", err)
	}

	return templates, nil
}

func (r *ProposalRepository) EnsureIndexes(ctx context.Context) error {
	collection := r.client.Database("freelanceX_proposals").Collection("proposals")

	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
		 Keys: bson.D{{Key: "client_id", Value: 1}},
		 Options: options.Index().SetName("client_id_index"),
		},
		{
		 Keys: bson.D{{Key: "freelancer_id", Value: 1}},
		 Options: options.Index().SetName("freelancer_id_index"),
		},
		{
		 Keys: bson.D{{Key: "status", Value: 1}},
		 Options: options.Index().SetName("status_index"),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}
