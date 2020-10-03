package weather

import (
	"github.com/seventy-two/evelyn/database"
)

type LocationStorage struct {
	Db *database.Database
}

func (s *LocationStorage) getLocation(userID string) (string, error) {
	cunt, err := s.Db.GetCunt(userID)
	if err != nil {
		return "", err
	}
	return cunt.Info.Location, nil
}

func (s *LocationStorage) setLocation(userID, location string) error {
	if err := s.Db.UpdateCuntInfo(userID, &database.CuntInfo{Location: location}); err != nil {
		return err
	}
	return nil
}
