package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/models"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

func PongHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {
	log := logger.Get(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "pong"}); err != nil {
		log.WithError(err).Error("Failed to encode JSON response")
		return err
	}

	return nil
}

func getUrlFilter(urlRawQuery string) string {
	log := logger.Get(context.Background())
	params, err := url.ParseQuery(urlRawQuery)
	if err != nil {
		log.WithError(err).Error("Failed to parse query parameters")
		return ""
	}

	var filter strings.Builder
	filter.Grow(len(urlRawQuery))

	for key, values := range params {
		if len(values) > 0 {
			filter.WriteString(key + ":" + values[0] + ",")
		}
	}
	return strings.TrimRight(filter.String(), ",")
}

func GetRepoHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {
	log := logger.Get(r.Context())
	w.Header().Set("Content-Type", "application/json")

	filter := getUrlFilter(r.URL.RawQuery)
	if filter == "" {
		log.Warn("No valid filters provided")
	}

	log.Infof("Applying filter: %s", filter)

	client := github.NewClient(nil)

	now := time.Now().Format(time.RFC3339)

	repos, err := fetchRepositories(client, filter, now, log)
	if err != nil {
		http.Error(w, "Failed to fetch repositories", http.StatusInternalServerError)
		return err
	}

	if repos == nil || repos.Repositories == nil {
		log.Error("No repositories found or response is nil")
		http.Error(w, "No repositories found", http.StatusNotFound)
		return nil
	}

	if err := json.NewEncoder(w).Encode(repos.Repositories); err != nil {
		log.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Failed to encode repositories", http.StatusInternalServerError)
		return err
	}

	return nil
}

func GetStatsHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) error {

	log := logger.Get(r.Context())
	w.Header().Set("Content-Type", "application/json")
	client := github.NewClient(nil)
	filter := getUrlFilter(r.URL.RawQuery)
	t := time.Now().Format(time.RFC3339)

	repos, err := fetchRepositories(client, filter, t, log)
	if err != nil {
		http.Error(w, "Failed to fetch repositories", http.StatusInternalServerError)
		return err
	}

	repoStats := processRepositories(client, repos, log)

	return encodeAndSendResponse(w, repoStats, log)
}

func fetchRepositories(client *github.Client, filter, t string, log logrus.FieldLogger) (*github.RepositoriesSearchResult, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	repos, _, err := client.Search.Repositories(context.Background(), filter+"created:<"+t, opt)
	if err != nil {
		log.WithError(err).Error("Failed to fetch repositories from GitHub")
		return nil, err
	}

	return repos, nil
}

func processRepositories(client *github.Client, repos *github.RepositoriesSearchResult, log logrus.FieldLogger) []models.Repo {
	var wg sync.WaitGroup
	repoStats := make([]models.Repo, len(repos.Repositories))
	mu := &sync.Mutex{}

	for i, repo := range repos.Repositories {
		wg.Add(1)
		go func(i int, repo github.Repository) {
			defer wg.Done()
			repoStat := processRepository(client, repo, log)

			mu.Lock()
			repoStats[i] = repoStat
			mu.Unlock()
		}(i, repo)
	}

	wg.Wait()
	return repoStats
}

func processRepository(client *github.Client, repo github.Repository, log logrus.FieldLogger) models.Repo {
	languages, _, err := client.Repositories.ListLanguages(context.Background(), repo.GetOwner().GetLogin(), repo.GetName())
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch languages for repository: %s", repo.GetFullName())
		return models.Repo{}
	}

	languageStats := make(map[string]models.Language)
	for lang, bytes := range languages {
		languageStats[lang] = models.Language{Bytes: bytes}
	}

	return models.Repo{
		FullName:   repo.FullName,
		Owner:      repo.Owner.Login,
		Repository: repo.Name,
		Languages:  languageStats,
	}
}

func encodeAndSendResponse(w http.ResponseWriter, repoStats []models.Repo, log logrus.FieldLogger) error {
	if err := json.NewEncoder(w).Encode(repoStats); err != nil {
		log.WithError(err).Error("Failed to encode JSON response")
		http.Error(w, "Failed to encode repositories", http.StatusInternalServerError)
		return err
	}
	return nil
}
