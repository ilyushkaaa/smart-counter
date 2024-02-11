package repo

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"smart-counter/internal/entity"
	errorapp "smart-counter/internal/errors"
)

type ComputingResourceRepo interface {
	GetComputingResourceForTask() (*entity.ComputingResource, error)
	GetComputingResources() ([]*entity.ComputingResource, error)
	DeleteComputingResource(cr *entity.ComputingResource) (bool, error)
	ChangeComputingResource(compRes *entity.ComputingResource, status, piingTime string) (*entity.ComputingResource, error)
}

type ComputingResourceMemoryRepo struct {
	computingResources          []*entity.ComputingResource
	db                          *sql.DB
	currentComputingResourceNum int
}

func NewComputingResourceMemoryRepo(db *sql.DB, computingResources []*entity.ComputingResource) *ComputingResourceMemoryRepo {
	return &ComputingResourceMemoryRepo{
		computingResources:          computingResources,
		currentComputingResourceNum: 0,
		db:                          db,
	}
}

func (cr *ComputingResourceMemoryRepo) GetComputingResources() ([]*entity.ComputingResource, error) {
	return cr.computingResources, nil
}

func (cr *ComputingResourceMemoryRepo) GetComputingResourceForTask() (*entity.ComputingResource, error) {
	for i := 0; i < len(cr.computingResources); i++ {
		currentResource := cr.computingResources[cr.currentComputingResourceNum]
		if cr.currentComputingResourceNum < len(cr.computingResources)-1 {
			cr.currentComputingResourceNum++
		} else {
			cr.currentComputingResourceNum = 0
		}
		if currentResource.Status == "ok" {
			return currentResource, nil
		}
	}
	return nil, errorapp.ErrorNoComputingResources
}

func (cr *ComputingResourceMemoryRepo) DeleteComputingResource(compRes *entity.ComputingResource) (bool, error) {
	for i := 0; i < len(cr.computingResources); i++ {
		if cr.computingResources[i] == compRes {
			cr.computingResources = append(cr.computingResources[:i], cr.computingResources[i+1:]...)
			return true, nil
		}
	}

	return false, nil
}

func (cr *ComputingResourceMemoryRepo) ChangeComputingResource(compRes *entity.ComputingResource, status, pingTime string) (*entity.ComputingResource, error) {
	for i := 0; i < len(cr.computingResources); i++ {
		if cr.computingResources[i] == compRes {
			cr.computingResources[i].Status = status
			cr.computingResources[i].LastPingTime = pingTime
			_, err := cr.db.Exec(
				"UPDATE computing_resource SET status = ?, last_ping_time = ? where id = ?",
				status,
				pingTime,
				compRes.ID,
			)
			if err != nil {
				return nil, err
			}
			return cr.computingResources[i], nil
		}
	}
	return nil, nil
}
