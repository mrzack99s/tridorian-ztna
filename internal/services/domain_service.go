package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"tridorian-ztna/internal/models"

	"gorm.io/gorm"
)

type DomainService struct {
	db *gorm.DB
}

func NewDomainService(db *gorm.DB) *DomainService {
	return &DomainService{db: db}
}

// GenerateVerificationToken generates a unique token for domain verification
func (s *DomainService) GenerateVerificationToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return "tridorian-verification=" + hex.EncodeToString(b)
}

// VerifyDomain checks if the domain has the correct TXT record
func (s *DomainService) VerifyDomain(ctx context.Context, customDomain *models.CustomDomain) (bool, error) {
	if customDomain.VerificationToken == "" {
		return false, fmt.Errorf("verification token is empty")
	}

	// 1. Check subdomain: tridorian-challenge.<domain>
	challengeHost := fmt.Sprintf("_tridorian-challenge.%s", customDomain.Domain)
	if s.checkTXTRecord(challengeHost, customDomain.VerificationToken) {
		return s.markVerified(customDomain)
	}

	log.Printf("Challenge TXT record not found for %s", challengeHost)

	// 2. Fallback: Check root domain
	if s.checkTXTRecord(customDomain.Domain, customDomain.VerificationToken) {
		return s.markVerified(customDomain)
	}

	log.Printf("Root TXT record not found for %s", customDomain.Domain)

	return false, nil
}

func (s *DomainService) checkTXTRecord(host, expectedToken string) bool {
	txtRecords, err := net.LookupTXT(host)
	if err != nil {
		return false
	}

	for _, record := range txtRecords {
		if strings.TrimSpace(record) == expectedToken {
			return true
		}
	}
	return false
}

func (s *DomainService) markVerified(customDomain *models.CustomDomain) (bool, error) {
	now := time.Now()
	customDomain.IsVerified = true
	customDomain.VerifiedAt = &now
	if err := s.db.Save(customDomain).Error; err != nil {
		return false, fmt.Errorf("failed to update domain status: %w", err)
	}
	return true, nil
}
