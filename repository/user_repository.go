package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"user-microservice-golang/config"
	apperrors "user-microservice-golang/errors"
	"user-microservice-golang/interfaces"
	"user-microservice-golang/model"
)

// userRepository is the MongoDB implementation of interfaces.UserRepository
type userRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewUserRepository constructs a userRepository and ensures required indexes
func NewUserRepository(mc *config.MongoClient, logger *zap.Logger) interfaces.UserRepository {
	col := mc.Collection(model.User{}.CollectionName())
	repo := &userRepository{collection: col, logger: logger}
	repo.ensureIndexes()
	return repo
}

// ensureIndexes creates the unique email index on startup
func (r *userRepository) ensureIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		r.logger.Warn("failed to create indexes", zap.Error(err))
	}
}

// baseFilter returns a filter that excludes soft-deleted documents
func baseFilter(extra bson.M) bson.M {
	f := bson.M{"deleted_at": bson.M{"$exists": false}}
	for k, v := range extra {
		f[k] = v
	}
	return f
}

// ─── Interface implementation ─────────────────────────────────────────────────

func (r *userRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	user.ID = primitive.NewObjectID()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	if _, err := r.collection.InsertOne(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, apperrors.ErrEmailAlreadyExists
		}
		r.logger.Error("create user failed", zap.Error(err))
		return nil, apperrors.ErrInternal
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	var user model.User
	err = r.collection.FindOne(ctx, baseFilter(bson.M{"_id": oid})).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrUserNotFound
		}
		r.logger.Error("find user by id failed", zap.String("id", id), zap.Error(err))
		return nil, apperrors.ErrInternal
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, baseFilter(bson.M{"email": email})).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrUserNotFound
		}
		r.logger.Error("find user by email failed", zap.String("email", email), zap.Error(err))
		return nil, apperrors.ErrInternal
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, page, limit int) ([]*model.User, int64, error) {
	filter := baseFilter(bson.M{})

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("count users failed", zap.Error(err))
		return nil, 0, apperrors.ErrInternal
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("find all users failed", zap.Error(err))
		return nil, 0, apperrors.ErrInternal
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("decode users failed", zap.Error(err))
		return nil, 0, apperrors.ErrInternal
	}
	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	updates["updated_at"] = time.Now().UTC()

	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated model.User
	err = r.collection.FindOneAndUpdate(
		ctx,
		baseFilter(bson.M{"_id": oid}),
		bson.M{"$set": updates},
		opts,
	).Decode(&updated)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrUserNotFound
		}
		r.logger.Error("update user failed", zap.String("id", id), zap.Error(err))
		return nil, apperrors.ErrInternal
	}
	return &updated, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	_, err := r.Update(ctx, id, map[string]interface{}{"password": hashedPassword})
	return err
}

func (r *userRepository) UpdateStatus(ctx context.Context, id string, status model.UserStatus) error {
	_, err := r.Update(ctx, id, map[string]interface{}{"status": string(status)})
	return err
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	now := time.Now().UTC()
	res, err := r.collection.UpdateOne(
		ctx,
		baseFilter(bson.M{"_id": oid}),
		bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}},
	)
	if err != nil {
		r.logger.Error("soft delete user failed", zap.String("id", id), zap.Error(err))
		return apperrors.ErrInternal
	}
	if res.MatchedCount == 0 {
		return apperrors.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, baseFilter(bson.M{"email": email}))
	if err != nil {
		r.logger.Error("exists by email failed", zap.String("email", email), zap.Error(err))
		return false, apperrors.ErrInternal
	}
	return count > 0, nil
}
