// Copyright (c) 2021 - 2024, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/autobrr/autobrr/internal/domain"

	"github.com/stretchr/testify/assert"
)

func getMockRelease() *domain.Release {
	return &domain.Release{
		FilterStatus: domain.ReleaseStatusFilterApproved,
		Rejections:   []string{"test", "not-a-match"},
		Indexer: domain.IndexerMinimal{
			ID:         0,
			Name:       "BTN",
			Identifier: "btn",
		},
		FilterName:     "ExampleFilter",
		Protocol:       domain.ReleaseProtocolTorrent,
		Implementation: domain.ReleaseImplementationIRC,
		Timestamp:      time.Now(),
		InfoURL:        "https://example.com/info",
		DownloadURL:    "https://example.com/download",
		GroupID:        "group123",
		TorrentID:      "torrent123",
		TorrentName:    "Example.Torrent.Name",
		Size:           123456789,
		Title:          "Example Title",
		Category:       "Movie",
		Season:         1,
		Episode:        2,
		Year:           2023,
		Resolution:     "1080p",
		Source:         "BluRay",
		Codec:          []string{"H.264", "AAC"},
		Container:      "MKV",
		HDR:            []string{"HDR10", "Dolby Vision"},
		Group:          "ExampleGroup",
		Proper:         true,
		Repack:         false,
		Website:        "https://example.com",
		Type:           "Movie",
		Origin:         "P2P",
		Tags:           []string{"Action", "Adventure"},
		Uploader:       "john_doe",
		PreTime:        "10m",
		FilterID:       1,
	}
}

func getMockReleaseActionStatus() *domain.ReleaseActionStatus {
	return &domain.ReleaseActionStatus{
		ID:         0,
		Status:     domain.ReleasePushStatusApproved,
		Action:     "okay",
		ActionID:   10,
		Type:       domain.ActionTypeTest,
		Client:     "qbitorrent",
		Filter:     "Test filter",
		FilterID:   0,
		Rejections: []string{"one rejection", "two rejections"},
		ReleaseID:  0,
		Timestamp:  time.Now(),
	}
}

func TestReleaseRepo_Store(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("StoreReleaseActionStatus_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			// Execute
			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Verify
			assert.NotEqual(t, int64(0), mockData.ID)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_StoreReleaseActionStatus(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("StoreReleaseActionStatus_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			// Execute
			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Verify
			assert.NotEqual(t, int64(0), releaseActionMockData.ID)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_Find(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		//actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		//releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("FindReleases_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			// Execute
			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)

			// Search with query params
			queryParams := domain.ReleaseQueryParams{
				Limit:  10,
				Offset: 0,
				Sort: map[string]string{
					"Timestamp": "asc",
				},
				Search: "",
			}

			resp, err := repo.Find(context.Background(), queryParams)

			// Verify
			assert.NotNil(t, resp)
			assert.NotEqual(t, int64(0), resp.TotalCount)
			assert.True(t, resp.NextCursor >= 0)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_FindRecent(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		//actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		//releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("FindRecent_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			// Execute
			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)

			resp, err := repo.Find(context.Background(), domain.ReleaseQueryParams{Limit: 10})

			// Verify
			assert.NotNil(t, resp.Data)
			assert.Lenf(t, resp.Data, 1, "Expected 1 release, got %d", len(resp.Data))

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_GetIndexerOptions(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("GetIndexerOptions_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Execute
			options, err := repo.GetIndexerOptions(context.Background())

			// Verify
			assert.NotNil(t, options)
			assert.Len(t, options, 1)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_GetActionStatusByReleaseID(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("GetActionStatusByReleaseID_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Execute
			actionStatus, err := repo.GetActionStatus(context.Background(), &domain.GetReleaseActionStatusRequest{Id: int(releaseActionMockData.ID)})

			// Verify
			assert.NoError(t, err)
			assert.NotNil(t, actionStatus)
			assert.Equal(t, releaseActionMockData.ID, actionStatus.ID)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_Get(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("Get_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Execute
			release, err := repo.Get(context.Background(), &domain.GetReleaseRequest{Id: int(mockData.ID)})

			// Verify
			assert.NoError(t, err)
			assert.NotNil(t, release)
			assert.Equal(t, mockData.ID, release.ID)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_Stats(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("Stats_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Execute
			stats, err := repo.Stats(context.Background())

			// Verify
			assert.NoError(t, err)
			assert.NotNil(t, stats)
			assert.Equal(t, int64(1), stats.PushApprovedCount)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_Delete(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("Delete_Succeeds [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			// Execute
			err = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})

			// Verify
			assert.NoError(t, err)

			// Cleanup
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func TestReleaseRepo_CheckSmartEpisodeCanDownloadShow(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		repo := NewReleaseRepo(log, db)

		mockData := getMockRelease()
		releaseActionMockData := getMockReleaseActionStatus()
		actionMockData := getMockAction()

		t.Run(fmt.Sprintf("Check_Smart_Episode_Can_Download [%s]", dbType), func(t *testing.T) {
			// Setup
			mock := getMockDownloadClient()
			err := downloadClientRepo.Store(context.Background(), &mock)
			assert.NoError(t, err)
			assert.NotNil(t, mock)

			err = filterRepo.Store(context.Background(), getMockFilter())
			assert.NoError(t, err)

			createdFilters, err := filterRepo.ListFilters(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, createdFilters)

			actionMockData.FilterID = createdFilters[0].ID
			actionMockData.ClientID = mock.ID
			mockData.FilterID = createdFilters[0].ID

			err = repo.Store(context.Background(), mockData)
			assert.NoError(t, err)
			err = actionRepo.Store(context.Background(), actionMockData)
			assert.NoError(t, err)

			releaseActionMockData.ReleaseID = mockData.ID
			releaseActionMockData.ActionID = int64(actionMockData.ID)
			releaseActionMockData.FilterID = int64(createdFilters[0].ID)

			err = repo.StoreReleaseActionStatus(context.Background(), releaseActionMockData)
			assert.NoError(t, err)

			params := &domain.SmartEpisodeParams{
				Title:   "Example.Torrent.Name",
				Season:  1,
				Episode: 2,
				Year:    0,
				Month:   0,
				Day:     0,
			}

			// Execute
			canDownload, err := repo.CheckSmartEpisodeCanDownload(context.Background(), params)

			// Verify
			assert.NoError(t, err)
			assert.True(t, canDownload)

			// Cleanup
			_ = repo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
			_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMockData.ID})
			_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
			_ = downloadClientRepo.Delete(context.Background(), mock.ID)
		})
	}
}

func getMockDuplicateReleaseProfileTV() *domain.DuplicateReleaseProfile {
	return &domain.DuplicateReleaseProfile{
		ID:          0,
		Name:        "TV",
		Protocol:    false,
		ReleaseName: false,
		Exact:       false,
		Title:       true,
		Year:        false,
		Month:       false,
		Day:         false,
		Source:      false,
		Resolution:  false,
		Codec:       false,
		Container:   false,
		HDR:         false,
		Audio:       false,
		Group:       false,
		Season:      true,
		Episode:     true,
		Website:     false,
		Proper:      false,
		Repack:      false,
	}
}

func getMockDuplicateReleaseProfileTVDaily() *domain.DuplicateReleaseProfile {
	return &domain.DuplicateReleaseProfile{
		ID:          0,
		Name:        "TV",
		Protocol:    false,
		ReleaseName: false,
		Exact:       false,
		Title:       true,
		Year:        true,
		Month:       true,
		Day:         true,
		Source:      false,
		Resolution:  false,
		Codec:       false,
		Container:   false,
		HDR:         false,
		Audio:       false,
		Group:       false,
		Season:      false,
		Episode:     false,
		Website:     false,
		Proper:      false,
		Repack:      false,
	}
}

func getMockFilterDuplicates() *domain.Filter {
	return &domain.Filter{
		Name:                 "New Filter",
		Enabled:              true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		MinSize:              "10mb",
		MaxSize:              "20mb",
		Delay:                60,
		Priority:             1,
		MaxDownloads:         100,
		MaxDownloadsUnit:     domain.FilterMaxDownloadsHour,
		MatchReleases:        "BRRip",
		ExceptReleases:       "BRRip",
		UseRegex:             false,
		MatchReleaseGroups:   "AMIABLE",
		ExceptReleaseGroups:  "NTb",
		Scene:                false,
		Origins:              nil,
		ExceptOrigins:        nil,
		Bonus:                nil,
		Freeleech:            false,
		FreeleechPercent:     "100%",
		SmartEpisode:         false,
		Shows:                "Is It Wrong to Try to Pick Up Girls in a Dungeon?",
		Seasons:              "4",
		Episodes:             "500",
		Resolutions:          []string{"1080p"},
		Codecs:               []string{"x264"},
		Sources:              []string{"BluRay"},
		Containers:           []string{"mkv"},
		MatchHDR:             []string{"HDR10"},
		ExceptHDR:            []string{"HDR10"},
		MatchOther:           []string{"Atmos"},
		ExceptOther:          []string{"Atmos"},
		Years:                "2023",
		Months:               "",
		Days:                 "",
		Artists:              "",
		Albums:               "",
		MatchReleaseTypes:    []string{"Remux"},
		ExceptReleaseTypes:   "Remux",
		Formats:              []string{"FLAC"},
		Quality:              []string{"Lossless"},
		Media:                []string{"CD"},
		PerfectFlac:          true,
		Cue:                  true,
		Log:                  true,
		LogScore:             100,
		MatchCategories:      "Anime",
		ExceptCategories:     "Anime",
		MatchUploaders:       "SubsPlease",
		ExceptUploaders:      "SubsPlease",
		MatchLanguage:        []string{"English", "Japanese"},
		ExceptLanguage:       []string{"English", "Japanese"},
		Tags:                 "Anime, x264",
		ExceptTags:           "Anime, x264",
		TagsAny:              "Anime, x264",
		ExceptTagsAny:        "Anime, x264",
		TagsMatchLogic:       "AND",
		ExceptTagsMatchLogic: "AND",
		MatchReleaseTags:     "Anime, x264",
		ExceptReleaseTags:    "Anime, x264",
		UseRegexReleaseTags:  true,
		MatchDescription:     "Anime, x264",
		ExceptDescription:    "Anime, x264",
		UseRegexDescription:  true,
	}
}

func TestReleaseRepo_CheckIsDuplicateRelease(t *testing.T) {
	for dbType, db := range testDBs {
		log := setupLoggerForTest()

		downloadClientRepo := NewDownloadClientRepo(log, db)
		filterRepo := NewFilterRepo(log, db)
		actionRepo := NewActionRepo(log, db, downloadClientRepo)
		releaseRepo := NewReleaseRepo(log, db)

		// reset
		//db.handler.Exec("DELETE FROM release")
		//db.handler.Exec("DELETE FROM action")
		//db.handler.Exec("DELETE FROM release_action_status")

		mockIndexer := domain.IndexerMinimal{ID: 0, Name: "Mock", Identifier: "mock", IdentifierExternal: "Mock"}
		actionMock := &domain.Action{Name: "Test", Type: domain.ActionTypeTest, Enabled: true}
		filterMock := getMockFilterDuplicates()

		// Setup
		err := filterRepo.Store(context.Background(), filterMock)
		assert.NoError(t, err)

		createdFilters, err := filterRepo.ListFilters(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, createdFilters)

		actionMock.FilterID = filterMock.ID

		err = actionRepo.Store(context.Background(), actionMock)
		assert.NoError(t, err)

		type fields struct {
			releaseTitles []string
			releaseTitle  string
			profile       *domain.DuplicateReleaseProfile
		}

		tests := []struct {
			name        string
			fields      fields
			isDuplicate bool
		}{
			{
				name: "1",
				fields: fields{
					releaseTitles: []string{
						"Inkheart 2008 BluRay 1080p DD5.1 x264-BADGROUP",
					},
					releaseTitle: "Inkheart 2008 BluRay 1080p DD5.1 x264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "2",
				fields: fields{
					releaseTitles: []string{
						"That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP",
						"That.Movie.2023.BluRay.720p.x265.DTS-HD-GROUP",
						"That.Movie.2023.WEB.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Source: true, Resolution: true},
				},
				isDuplicate: true,
			},
			{
				name: "3",
				fields: fields{
					releaseTitles: []string{
						"That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP",
						"That.Movie.2023.BluRay.720p.x265.DTS-HD-GROUP",
						"That.Movie.2023.WEB.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Source: true, Codec: true, Resolution: true},
				},
				isDuplicate: true,
			},
			{
				name: "4",
				fields: fields{
					releaseTitles: []string{
						"That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP",
						"That.Movie.2023.BluRay.720p.x265.DTS-HD-GROUP",
						"That.Movie.2023.WEB.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Movie.2023.BluRay.2160p.x265.DTS-HD-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Source: true, Codec: true, Resolution: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "5",
				fields: fields{
					releaseTitles: []string{
						"That.Tv.Show.2023.S01E01.BluRay.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Tv.Show.2023.S01E01.BluRay.2160p.x265.DTS-HD-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Group: true},
				},
				isDuplicate: true,
			},
			{
				name: "6",
				fields: fields{
					releaseTitles: []string{
						"That.Tv.Show.2023.S01E01.BluRay.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Tv.Show.2023.S01E02.BluRay.2160p.x265.DTS-HD-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "7",
				fields: fields{
					releaseTitles: []string{
						"That.Tv.Show.2023.S01.BluRay.2160p.x265.DTS-HD-GROUP",
					},
					releaseTitle: "That.Tv.Show.2023.S01.BluRay.2160p.x265.DTS-HD-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Group: true},
				},
				isDuplicate: true,
			},
			{
				name: "8",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 SDR H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 SDR H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, Group: true},
				},
				isDuplicate: true,
			},
			{
				name: "9",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.HULU.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 SDR H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "10",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, HDR: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "11",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
						"The Best Show 2020 S04E10 1080p amzn web-dl ddp 5.1 hdr dv h.264-group",
					},
					releaseTitle: "The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, HDR: true},
				},
				isDuplicate: true,
			},
			{
				name: "12",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 1080p AMZN WEB-DL DDP 5.1 DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, HDR: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "13",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 Episode Title 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, SubTitle: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, HDR: true, Group: true},
				},
				isDuplicate: false,
			},
			{
				name: "14",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.Episode.Title.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 Episode Title 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, SubTitle: true, Year: true, Season: true, Episode: true, Source: true, Codec: true, Resolution: true, Website: true, HDR: true, Group: true},
				},
				isDuplicate: true,
			},
			{
				name: "15",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.Episode.Title.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 Episode Title 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, SubTitle: true, Season: true, Episode: true, HDR: true},
				},
				isDuplicate: true,
			},
			{
				name: "16",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.Episode.Title.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E11 Episode Title 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, SubTitle: true, Season: true, Episode: true, HDR: true},
				},
				isDuplicate: false,
			},
			{
				name: "17",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.Episode.Title.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP",
					},
					releaseTitle: "The Best Show 2020 S04E10 Episode Title REPACK 1080p AMZN WEB-DL DDP 5.1 HDR DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, SubTitle: true, Season: true, Episode: true, HDR: true},
				},
				isDuplicate: true,
			},
			{
				name: "18",
				fields: fields{
					releaseTitles: []string{
						"The Best Show 2020 S04E10 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The.Best.Show.2020.S04E10.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The.Best.Show.2020.S04E10.Episode.Title.REPACK.1080p.AMZN.WEB-DL.DDP.5.1.HDR.DV.H.264-GROUP1",
					},
					releaseTitle: "The Best Show 2020 S04E10 Episode Title REPACK 1080p AMZN WEB-DL DDP 5.1 DV H.264-GROUP",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Season: true, Episode: true, Repack: true},
				},
				isDuplicate: false,
			},
			{
				name: "19",
				fields: fields{
					releaseTitles: []string{
						"The Daily Show 2024-09-21 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The Daily Show 2024-09-21.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The Daily Show 2024-09-21.Guest.1080p.AMZN.WEB-DL.DDP.5.1.H.264-GROUP1",
					},
					releaseTitle: "The Daily Show 2024-09-21.Other.Guest.1080p.AMZN.WEB-DL.DDP.5.1.H.264-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Season: true, Episode: true, Year: true, Month: true, Day: true},
				},
				isDuplicate: true,
			},
			{
				name: "20",
				fields: fields{
					releaseTitles: []string{
						"The Daily Show 2024-09-21 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The Daily Show 2024-09-21.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The Daily Show 2024-09-21.Guest.1080p.AMZN.WEB-DL.DDP.5.1.H.264-GROUP1",
					},
					releaseTitle: "The Daily Show 2024-09-21 Other Guest 1080p AMZN WEB-DL DDP 5.1 H.264-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Season: true, Episode: true, Year: true, Month: true, Day: true, SubTitle: true},
				},
				isDuplicate: false,
			},
			{
				name: "21",
				fields: fields{
					releaseTitles: []string{
						"The Daily Show 2024-09-21 1080p HULU WEB-DL DDP 5.1 SDR H.264-GROUP",
						"The Daily Show 2024-09-21.1080p.AMZN.WEB-DL.DDP.5.1.SDR.H.264-GROUP",
						"The Daily Show 2024-09-21.Guest.1080p.AMZN.WEB-DL.DDP.5.1.H.264-GROUP1",
					},
					releaseTitle: "The Daily Show 2024-09-22 Other Guest 1080p AMZN WEB-DL DDP 5.1 H.264-GROUP1",
					profile:      &domain.DuplicateReleaseProfile{Title: true, Season: true, Episode: true, Year: true, Month: true, Day: true, SubTitle: true},
				},
				isDuplicate: false,
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("Check_Is_Duplicate_Release %s [%s]", tt.name, dbType), func(t *testing.T) {
				ctx := context.Background()

				// Setup
				for _, rel := range tt.fields.releaseTitles {
					mockRel := domain.NewRelease(mockIndexer)
					mockRel.ParseString(rel)

					mockRel.FilterID = filterMock.ID

					err = releaseRepo.Store(ctx, mockRel)
					assert.NoError(t, err)

					ras := &domain.ReleaseActionStatus{
						ID:         0,
						Status:     domain.ReleasePushStatusApproved,
						Action:     "test",
						ActionID:   int64(actionMock.ID),
						Type:       domain.ActionTypeTest,
						Client:     "",
						Filter:     "Test filter",
						FilterID:   int64(filterMock.ID),
						Rejections: []string{},
						ReleaseID:  mockRel.ID,
						Timestamp:  time.Now(),
					}

					err = releaseRepo.StoreReleaseActionStatus(ctx, ras)
					assert.NoError(t, err)
				}

				releases, err := releaseRepo.Find(ctx, domain.ReleaseQueryParams{})
				assert.NoError(t, err)
				assert.Len(t, releases.Data, len(tt.fields.releaseTitles))

				compareRel := domain.NewRelease(mockIndexer)
				compareRel.ParseString(tt.fields.releaseTitle)

				// Execute
				isDuplicate, err := releaseRepo.CheckIsDuplicateRelease(ctx, tt.fields.profile, compareRel.Normalized())

				// Verify
				assert.NoError(t, err)
				assert.Equal(t, tt.isDuplicate, isDuplicate)

				// Cleanup
				_ = releaseRepo.Delete(ctx, &domain.DeleteReleaseRequest{OlderThan: 0})
			})
		}

		// Cleanup
		//_ = releaseRepo.Delete(context.Background(), &domain.DeleteReleaseRequest{OlderThan: 0})
		_ = actionRepo.Delete(context.Background(), &domain.DeleteActionRequest{ActionId: actionMock.ID})
		_ = filterRepo.Delete(context.Background(), createdFilters[0].ID)
	}
}
