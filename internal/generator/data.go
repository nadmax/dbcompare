package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nadmax/dbcompare/internal/models"
)

var (
	firstNames = []string{
		"James", "Mary", "John", "Patricia", "Robert", "Jennifer", "Michael", "Linda",
		"William", "Barbara", "David", "Elizabeth", "Richard", "Susan", "Joseph", "Jessica",
		"Thomas", "Sarah", "Charles", "Karen", "Christopher", "Nancy", "Daniel", "Lisa",
	}

	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
		"Taylor", "Moore", "Jackson", "Martin", "Lee", "Thompson", "White", "Harris",
	}

	domains = []string{
		"example.com", "test.com", "demo.com", "sample.org", "email.net",
		"mail.com", "inbox.io", "webmail.com", "contact.biz", "corp.com",
	}

	descriptions = []string{
		"Experienced professional with strong background in",
		"Dedicated team member focused on",
		"Results-driven individual specializing in",
		"Innovative thinker passionate about",
		"Strategic planner with expertise in",
		"Dynamic professional committed to",
		"Detail-oriented specialist in",
		"Accomplished expert with knowledge of",
	}

	skills = []string{
		"project management", "data analysis", "software development", "customer service",
		"marketing strategy", "financial planning", "quality assurance", "business operations",
		"technical support", "product design", "team leadership", "process improvement",
	}
)

type Generator struct {
	rand *rand.Rand
}

func New(seed int64) *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func NewDefault() *Generator {
	return New(time.Now().UnixNano())
}

func (g *Generator) GenerateRecord(id int) models.TestRecord {
	firstName := firstNames[g.rand.Intn(len(firstNames))]
	lastName := lastNames[g.rand.Intn(len(lastNames))]

	return models.TestRecord{
		ID:          id,
		Name:        fmt.Sprintf("%s %s", firstName, lastName),
		Email:       g.generateEmail(firstName, lastName, id),
		Age:         g.rand.Intn(60) + 18,
		Balance:     float64(g.rand.Intn(100000)) / 100,
		CreatedAt:   time.Now().Add(-time.Duration(g.rand.Intn(365*24)) * time.Hour),
		Description: g.generateDescription(),
		IsActive:    g.rand.Float32() > 0.3,
	}
}

func (g *Generator) GenerateRecords(count int, startID int) []models.TestRecord {
	records := make([]models.TestRecord, count)
	for i := range count {
		records[i] = g.GenerateRecord(startID + i)
	}
	return records
}

func (g *Generator) GenerateBatch(batchSize int, batchNumber int) []models.TestRecord {
	startID := batchNumber * batchSize
	return g.GenerateRecords(batchSize, startID)
}

func (g *Generator) generateEmail(firstName, lastName string, id int) string {
	domain := domains[g.rand.Intn(len(domains))]

	formats := []string{
		"%s.%s@%s",
		"%s_%s@%s",
		"%s%d@%s",
		"%s.%s%d@%s",
	}

	format := formats[g.rand.Intn(len(formats))]

	switch format {
	case "%s.%s@%s":
		return fmt.Sprintf(format, firstName, lastName, domain)
	case "%s_%s@%s":
		return fmt.Sprintf(format, firstName, lastName, domain)
	case "%s%d@%s":
		return fmt.Sprintf(format, firstName, id, domain)
	case "%s.%s%d@%s":
		return fmt.Sprintf(format, firstName, lastName, id, domain)
	default:
		return fmt.Sprintf("%s.%s@%s", firstName, lastName, domain)
	}
}

func (g *Generator) generateDescription() string {
	desc := descriptions[g.rand.Intn(len(descriptions))]
	skill := skills[g.rand.Intn(len(skills))]

	extras := []string{
		" and delivering exceptional results.",
		" with proven track record.",
		" and continuous improvement.",
		" to drive organizational success.",
		" and collaborative problem-solving.",
	}

	extra := extras[g.rand.Intn(len(extras))]

	return desc + " " + skill + extra
}

func (g *Generator) GenerateRandomID(max int) int {
	if max <= 0 {
		return 1
	}
	return g.rand.Intn(max) + 1
}

func (g *Generator) GenerateRandomIDs(count, max int) []int {
	if count > max {
		count = max
	}

	ids := make(map[int]bool)
	result := make([]int, 0, count)

	for len(result) < count {
		id := g.GenerateRandomID(max)
		if !ids[id] {
			ids[id] = true
			result = append(result, id)
		}
	}

	return result
}

func (g *Generator) GenerateUpdateValue(fieldType string) any {
	switch fieldType {
	case "balance":
		return float64(g.rand.Intn(10000)) / 100
	case "age":
		return g.rand.Intn(60) + 18
	case "active":
		return g.rand.Float32() > 0.5
	case "name":
		return fmt.Sprintf("%s %s",
			firstNames[g.rand.Intn(len(firstNames))],
			lastNames[g.rand.Intn(len(lastNames))])
	default:
		return nil
	}
}
