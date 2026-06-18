package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/calli-machtia/backend/internal/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LessonRepo struct {
	coll *mongo.Collection
}

func NewLessonRepo(db *mongo.Database) *LessonRepo {
	if db == nil {
		return &LessonRepo{coll: nil}
	}
	return &LessonRepo{coll: db.Collection("modules")}
}

func (r *LessonRepo) GetModules(courseID string) ([]models.Module, error) {
	if r.coll == nil {
		return []models.Module{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"course_id": courseID})
	if err != nil {
		return nil, fmt.Errorf("get modules: %w", err)
	}
	defer cursor.Close(ctx)

	var modules []models.Module
	if err := cursor.All(ctx, &modules); err != nil {
		return nil, fmt.Errorf("decode modules: %w", err)
	}

	if modules == nil {
		modules = []models.Module{}
	}

	return modules, nil
}

func (r *LessonRepo) GetModule(moduleID string) (*models.Module, error) {
	if r.coll == nil {
		return nil, ErrNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var module models.Module
	err := r.coll.FindOne(ctx, bson.M{"_id": moduleID}).Decode(&module)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get module: %w", err)
	}
	return &module, nil
}

func (r *LessonRepo) CreateModule(module *models.Module) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	module.ID = uuid.New().String()
	module.CreatedAt = time.Now()
	module.UpdatedAt = time.Now()
	if module.Lessons == nil {
		module.Lessons = []models.Lesson{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.InsertOne(ctx, module)
	if err != nil {
		return fmt.Errorf("create module: %w", err)
	}
	return nil
}

func (r *LessonRepo) UpdateModule(module *models.Module) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	module.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": module.ID},
		bson.M{"$set": bson.M{
			"title":       module.Title,
			"description": module.Description,
			"order":       module.Order,
			"updated_at":  module.UpdatedAt,
		}},
	)
	if err != nil {
		return fmt.Errorf("update module: %w", err)
	}
	return nil
}

func (r *LessonRepo) DeleteModule(moduleID string) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": moduleID})
	if err != nil {
		return fmt.Errorf("delete module: %w", err)
	}
	return nil
}

func (r *LessonRepo) AddLesson(moduleID string, lesson *models.Lesson) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	lesson.ID = uuid.New().String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": moduleID},
		bson.M{"$push": bson.M{"lessons": lesson}, "$set": bson.M{"updated_at": time.Now()}},
	)
	if err != nil {
		return fmt.Errorf("add lesson: %w", err)
	}
	return nil
}

func (r *LessonRepo) UpdateLesson(moduleID, lessonID string, lesson *models.Lesson) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	lesson.ID = lessonID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": moduleID, "lessons.id": lessonID},
		bson.M{"$set": bson.M{
			"lessons.$.title":       lesson.Title,
			"lessons.$.description": lesson.Description,
			"lessons.$.content":     lesson.Content,
			"lessons.$.video_url":   lesson.VideoURL,
			"lessons.$.duration":    lesson.Duration,
			"lessons.$.order":       lesson.Order,
			"lessons.$.free":        lesson.Free,
			"updated_at":            time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("update lesson: %w", err)
	}
	return nil
}

func (r *LessonRepo) DeleteLesson(moduleID, lessonID string) error {
	if r.coll == nil {
		return fmt.Errorf("MongoDB no disponible")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": moduleID},
		bson.M{"$pull": bson.M{"lessons": bson.M{"id": lessonID}}, "$set": bson.M{"updated_at": time.Now()}},
	)
	if err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}
