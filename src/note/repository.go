package note

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Repository is the interface can be used in DI
type Repository interface {
	CreateNote(note Note) (Note, error)
	GetNotes() ([]Note, error)
	GetNoteById(id int) (Note, error)
	UpdateNoteById(note Note) error
	DeleteNoteById(id int) error
	DeleteAllNotes() error

	GetNotesByTag(tag string, page int, pageSize int) ([]Note, error)
	GetAllTags() ([]Tag, error)
}

// RepositoryImpl is the default implementation of Repository
type RepositoryImpl struct {
	DB *gorm.DB
}

// NewNewRepository is the constructor of RepositoryImpl
func NewRepository(db *gorm.DB) Repository {
	return &RepositoryImpl{
		DB: db,
	}
}

// CreateNote is to create a new Note
func (r *RepositoryImpl) CreateNote(note Note) (Note, error) {
	// log.Debug().Msgf("to be added %v", note)

	r.DB.Create(&note)

	return note, r.DB.Error
}

// CreateNote is to create a new Note
func (r *RepositoryImpl) UpdateNoteById(note Note) error {
	log.Debug().Uint("note", note.ID).Msg("Update note by id")

	r.DB.Save(&note)

	return r.DB.Error
}

// GetNotes Get all notes
// This is only used for perforamnce testing and it didn't join tables
func (r *RepositoryImpl) GetNotes() ([]Note, error) {
	var notes []Note
	r.DB.Find(&notes)

	return notes, r.DB.Error
}

// GetNotes Get all notes from db per page number and pagesize
func (r *RepositoryImpl) GetNotesWithPagination(page int, pageSize int) ([]Note, error) {
	var notes []Note
	r.DB.Scopes(Paginate(page, pageSize)).Find(&notes)

	return notes, r.DB.Error
}

// GetNotes Get all notes from db per page number and pagesize
func (r *RepositoryImpl) GetNotesByTag(tag string, page int, pageSize int) ([]Note, error) {
	log.Debug().Str("tag", tag).Msg("Get notes by tag")
	var notes []Note
	// TODO: There is a bug here that Preload tags with null pointer exception
	r.DB.
		// Debug().
		Joins("JOIN note_tags on note_tags.note_id = notes.id").
		Joins("JOIN tags on note_tags.tag_name = tags.name").
		// Preload("tags").
		Where("tags.name = ?", tag).
		Group("notes.id").
		Find(&notes)

	return notes, r.DB.Error
}

// GetNote by Id from db
func (r *RepositoryImpl) GetNoteById(id int) (Note, error) {
	var note Note
	log.Debug().Int("id", id).Msg("Retrieve note by id")
	r.DB.
		// Debug().
		Where("notes.id = ?", id).Preload("Tags").First(&note)
	return note, r.DB.Error
}

// DeleteNoteById remove a note by Id
func (r *RepositoryImpl) DeleteNoteById(id int) error {
	r.DB.Where("id = ?", id).Delete(&Note{})

	return r.DB.Error
}

// DeleteAllNotes is only for unit testing
func (r *RepositoryImpl) DeleteAllNotes() error {
	r.DB.Unscoped().Where("1=1").Delete(&Note{})
	r.DB.Unscoped().Where("1=1").Delete(&Tag{})

	return r.DB.Error
}

func (r *RepositoryImpl) GetTagsByName(tag string) ([]Tag, error) {
	var tags []Tag
	r.DB.Where("name = ?", tag).Find(&tags)
	return tags, r.DB.Error
}

// GetNotes Get all tags from db
func (r *RepositoryImpl) GetAllTags() ([]Tag, error) {
	var tags []Tag
	r.DB.Find(&tags)

	return tags, r.DB.Error
}

// Paginate The generic pagination function
// @Param	page - optional - page number of the query
// 			pageSize - optional - page size of the query
// @Success object array
// @Failure DB.error
func Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}