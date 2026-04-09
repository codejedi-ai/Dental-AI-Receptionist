// Package intentclassifier implements Module 2: Intent Classification
// for the Dental AI Receptionist system.
//
// It receives the output from Module 1 (lang_code + first_response) along with
// the patient's full first utterance, then determines the patient's intent
// and extracts relevant entities for downstream processing.
package intentclassifier

import (
	"log"
	"strings"
)

// ─── Intent Types ───────────────────────────────────────────────

// Intent represents the classified intent of the patient's utterance.
type Intent string

const (
	IntentBook       Intent = "book_appointment"
	IntentCancel     Intent = "cancel_appointment"
	IntentReschedule Intent = "reschedule_appointment"
	IntentAskHours   Intent = "ask_hours"
	IntentAskLocation Intent = "ask_location"
	IntentAskServices Intent = "ask_services"
	IntentEmergency  Intent = "emergency"
	IntentGreeting   Intent = "greeting"
	IntentUnclear    Intent = "unclear"
)

// ─── Classification Result ──────────────────────────────────────

// ClassificationResult is the structured output consumed by downstream modules.
type ClassificationResult struct {
	Module    string  `json:"module"`
	Intent    Intent  `json:"intent"`
	Confidence float64 `json:"confidence"` // 0.0 - 1.0
	LangCode  string  `json:"lang_code"`
	Entities  Entities `json:"entities"`
	NextModule string `json:"next_module"`
	Metadata  ClassificationMetadata `json:"metadata"`
}

// Entities holds extracted information from the utterance.
type Entities struct {
	Service        *string `json:"service,omitempty"`
	PreferredDate  *string `json:"preferred_date,omitempty"`
	PreferredTime  *string `json:"preferred_time,omitempty"`
	Dentist        *string `json:"dentist,omitempty"`
	AppointmentID  *string `json:"appointment_id,omitempty"`
	PatientName    *string `json:"patient_name,omitempty"`
	Reason         *string `json:"reason,omitempty"`
}

// ClassificationMetadata holds diagnostic info.
type ClassificationMetadata struct {
	InputSentence string   `json:"input_sentence"`
	MatchedRules  []string `json:"matched_rules"`
}

// ─── Confidence Thresholds ──────────────────────────────────────

const (
	ConfidenceExactMatch  = 0.95
	ConfidenceStrongMatch = 0.85
	ConfidenceWeakMatch   = 0.60
	ConfidenceBelowThreshold = 0.40
)

// ConfidenceThreshold is the minimum confidence to act without clarification.
const ConfidenceThreshold = 0.70

// ─── Service Keywords ───────────────────────────────────────────

// serviceKeywords maps service keywords to canonical service names.
var serviceKeywords = map[string]string{
	"cleaning":        "Cleaning",
	"clean":           "Cleaning",
	"checkup":         "Checkup",
	"check up":        "Checkup",
	"exam":            "Checkup",
	"examination":     "Checkup",
	"consultation":    "Consultation",
	"consult":         "Consultation",
	"whitening":       "Whitening",
	"whiten":          "Whitening",
	"bleaching":       "Whitening",
	"filling":         "Filling",
	"fill":            "Filling",
	"cavity":          "Filling",
	"crown":           "Crown",
	"cap":             "Crown",
	"bridge":          "Bridge",
	"root canal":      "Root Canal",
	"implant":         "Implant",
	"invisalign":      "Invisalign",
	"braces":          "Invisalign",
	"straighten":      "Invisalign",
	"pediatric":       "Pediatric",
	"child":           "Pediatric",
	"kids":            "Pediatric",
	"emergency":       "Emergency",
	"urgent":          "Emergency",
	"pain":            "Emergency",
	"toothache":       "Emergency",
	"broken":          "Emergency",
	"chipped":         "Emergency",
	"swelling":        "Emergency",
	"bleeding":        "Emergency",
}

// Chinese service keywords.
var serviceKeywordsZH = map[string]string{
	"洗牙":    "Cleaning",
	"检查":    "Checkup",
	"体检":    "Checkup",
	"咨询":    "Consultation",
	"美白":    "Whitening",
	"补牙":    "Filling",
	"牙冠":    "Crown",
	"牙桥":    "Bridge",
	"根管":    "Root Canal",
	"种植":    "Implant",
	"正畸":    "Invisalign",
	"儿童":    "Pediatric",
	"急诊":    "Emergency",
	"牙痛":    "Emergency",
	"痛":      "Emergency",
}

// ─── Intent Rule Definitions ────────────────────────────────────

// intentRule defines a single keyword-based classification rule.
type intentRule struct {
	Intent       Intent
	Keywords     []string // English keywords
	KeywordsZH   []string // Chinese keywords
	NegativeEN   []string // English keywords that disqualify this rule
	NegativeZH   []string // Chinese keywords that disqualify this rule
	Confidence   float64
	RuleName     string
}

// intentRules is ordered by specificity (most specific first).
var intentRules = []intentRule{
	{
		Intent:     IntentEmergency,
		Keywords:   []string{"emergency", "urgent", "severe pain", "toothache", "broken tooth", "chipped", "swelling", "bleeding", "knocked out"},
		KeywordsZH: []string{"急诊", "紧急", "很痛", "牙痛", "牙掉了", "出血", "肿"},
		Confidence: ConfidenceExactMatch,
		RuleName:   "emergency_keywords",
	},
	{
		Intent:     IntentCancel,
		Keywords:   []string{"cancel", "cancel appointment", "don't need", "no longer need", "remove my"},
		KeywordsZH: []string{"取消", "取消预约"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "cancellation_keywords",
	},
	{
		Intent:     IntentReschedule,
		Keywords:   []string{"reschedule", "move", "change", "postpone", "different time", "another time"},
		KeywordsZH: []string{"改期", "改时间", "换时间", "推迟"},
		NegativeEN: []string{"cancel"},
		NegativeZH: []string{"取消"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "reschedule_keywords",
	},
	{
		Intent:     IntentBook,
		Keywords:   []string{"book", "schedule", "make an appointment", "new appointment", "set up", "arrange", "reserve"},
		KeywordsZH: []string{"预约", "挂号", "安排"},
		NegativeEN: []string{"cancel", "reschedule", "move", "change", "postpone"},
		NegativeZH: []string{"取消", "改期", "改时间", "换时间", "推迟"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "booking_keywords",
	},
	{
		Intent:     IntentAskHours,
		Keywords:   []string{"hours", "open", "close", "when are you", "business hours", "working hours", "operating hours"},
		KeywordsZH: []string{"营业时间", "几点开门", "几点关门", "什么时候上班"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "hours_keywords",
	},
	{
		Intent:     IntentAskLocation,
		Keywords:   []string{"location", "address", "where", "directions", "how to get", "find you"},
		KeywordsZH: []string{"地址", "在哪里", "怎么走", "位置"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "location_keywords",
	},
	{
		Intent:     IntentAskServices,
		Keywords:   []string{"service", "do you offer", "do you do", "what services", "types of", "treatments"},
		KeywordsZH: []string{"服务", "有没有", "提供", "项目"},
		Confidence: ConfidenceStrongMatch,
		RuleName:   "services_keywords",
	},
	{
		Intent:     IntentGreeting,
		Keywords:   []string{"hello", "hi", "hey", "good morning", "good afternoon", "good evening", "greetings"},
		KeywordsZH: []string{"你好", "您好", "喂", "嗨"},
		Confidence: ConfidenceBelowThreshold,
		RuleName:   "greeting_keywords",
	},
}

// ─── Classify ───────────────────────────────────────────────────

// Classify analyzes the patient's utterance and returns a ClassificationResult.
//
// langCode should come from Module 1's output. The classifier uses it to
// weight English vs Chinese keyword matching, but will check both regardless.
func Classify(utterance string, langCode string) ClassificationResult {
	result := ClassificationResult{
		Module:     "intent_classification",
		LangCode:   langCode,
		Intent:     IntentUnclear,
		Confidence: 0.0,
		Entities:   Entities{},
		NextModule: "task_handler",
		Metadata: ClassificationMetadata{
			InputSentence: utterance,
			MatchedRules:  []string{},
		},
	}

	lower := strings.ToLower(strings.TrimSpace(utterance))
	if lower == "" {
		result.Confidence = ConfidenceBelowThreshold
		result.NextModule = "clarification_needed"
		log.Println("[IntentClassify] Empty utterance, classified as unclear")
		return result
	}

	// Check each rule in priority order
	var bestMatch intentRule
	var bestScore float64

	for _, rule := range intentRules {
		score := matchRule(lower, rule, langCode)
		if score > bestScore {
			bestScore = score
			bestMatch = rule
		}
	}

	if bestScore > 0 {
		result.Intent = bestMatch.Intent
		result.Confidence = bestScore
		result.Metadata.MatchedRules = append(result.Metadata.MatchedRules, bestMatch.RuleName)
	}

	// Extract entities regardless of intent
	result.Entities = extractEntities(lower, utterance, langCode)

	// Determine next module based on intent
	switch result.Intent {
	case IntentBook:
		result.NextModule = "appointment_booking"
	case IntentCancel:
		result.NextModule = "appointment_cancellation"
	case IntentReschedule:
		result.NextModule = "appointment_reschedule"
	case IntentAskHours, IntentAskLocation, IntentAskServices:
		result.NextModule = "knowledge_base"
	case IntentEmergency:
		result.NextModule = "emergency_handler"
	case IntentGreeting:
		result.NextModule = "greeting_handler"
	default:
		result.NextModule = "clarification_needed"
	}

	log.Printf("[IntentClassify] Input=%q intent=%s confidence=%.2f lang=%s",
		utterance, result.Intent, result.Confidence, langCode)

	return result
}

// matchRule scores how well the utterance matches an intent rule.
// Returns 0 if any negative keyword is found.
func matchRule(lower string, rule intentRule, langCode string) float64 {
	// Check negative keywords — if any match, disqualify this rule entirely
	for _, kw := range rule.NegativeEN {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return 0
		}
	}
	for _, kw := range rule.NegativeZH {
		if strings.Contains(lower, kw) {
			return 0
		}
	}

	maxScore := 0.0

	// Check English keywords
	for _, kw := range rule.Keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			score := rule.Confidence
			// Boost if multi-word phrase matches
			if strings.Count(kw, " ") > 0 {
				score = min(score+0.05, 1.0)
			}
			if score > maxScore {
				maxScore = score
			}
		}
	}

	// Check Chinese keywords
	for _, kw := range rule.KeywordsZH {
		if strings.Contains(lower, kw) || strings.Contains(lower, strings.ToLower(kw)) {
			score := rule.Confidence
			// Boost if lang is Chinese
			if langCode == "zh" {
				score = min(score+0.1, 1.0)
			}
			if score > maxScore {
				maxScore = score
			}
		}
	}

	return maxScore
}

// ─── Entity Extraction ──────────────────────────────────────────

// extractEntities pulls relevant data from the utterance.
func extractEntities(lower string, original string, langCode string) Entities {
	entities := Entities{}

	// Service extraction
	service := extractService(lower, langCode)
	if service != "" {
		entities.Service = &service
	}

	// Date extraction (simple pattern: YYYY-MM-DD, MM/DD/YYYY, "next Monday", etc.)
	date := extractDate(lower)
	if date != "" {
		entities.PreferredDate = &date
	}

	// Time extraction (simple pattern: HH:MM, "3pm", "afternoon", etc.)
	timeStr := extractTime(lower)
	if timeStr != "" {
		entities.PreferredTime = &timeStr
	}

	// Dentist name extraction
	dentist := extractDentist(lower, original)
	if dentist != "" {
		entities.Dentist = &dentist
	}

	// Reason extraction (for emergencies or rescheduling)
	if strings.Contains(lower, "toothache") || strings.Contains(lower, "pain") ||
		strings.Contains(lower, "牙痛") || strings.Contains(lower, "痛") {
		reason := "Dental pain / toothache"
		entities.Reason = &reason
	}

	return entities
}

// extractService finds a service name in the utterance.
func extractService(lower string, langCode string) string {
	// Check Chinese first if lang is Chinese
	if langCode == "zh" {
		for zh, canonical := range serviceKeywordsZH {
			if strings.Contains(lower, zh) {
				return canonical
			}
		}
	}
	// Then check English
	for kw, canonical := range serviceKeywords {
		if strings.Contains(lower, kw) {
			return canonical
		}
	}
	return ""
}

// extractDate finds a date reference in the utterance.
func extractDate(lower string) string {
	// Simple patterns — a production system would use a date parser
	datePatterns := []struct {
		pattern string
		result  string
	}{
		{"tomorrow", "tomorrow"},
		{"next monday", "next_monday"},
		{"next tuesday", "next_tuesday"},
		{"next wednesday", "next_wednesday"},
		{"next thursday", "next_thursday"},
		{"next friday", "next_friday"},
		{"next saturday", "next_saturday"},
		{"monday", "monday"},
		{"tuesday", "tuesday"},
		{"wednesday", "wednesday"},
		{"thursday", "thursday"},
		{"friday", "friday"},
		{"saturday", "saturday"},
		{"明天", "tomorrow"},
		{"下周一", "next_monday"},
		{"下周二", "next_tuesday"},
		{"下周三", "next_wednesday"},
		{"下周四", "next_thursday"},
		{"下周五", "next_friday"},
		{"下周六", "next_saturday"},
	}

	for _, p := range datePatterns {
		if strings.Contains(lower, p.pattern) {
			return p.result
		}
	}
	return ""
}

// extractTime finds a time reference in the utterance.
func extractTime(lower string) string {
	timePatterns := []struct {
		pattern string
		result  string
	}{
		{"morning", "morning"},
		{"afternoon", "afternoon"},
		{"evening", "evening"},
		{"上午", "morning"},
		{"下午", "afternoon"},
		{"晚上", "evening"},
	}

	for _, p := range timePatterns {
		if strings.Contains(lower, p.pattern) {
			return p.result
		}
	}
	return ""
}

// extractDentist finds a dentist name in the utterance.
func extractDentist(lower string, original string) string {
	dentists := []string{
		"Dr. Sarah Chen", "Dr. Chen", "Sarah Chen",
		"Dr. Michael Park", "Dr. Park", "Michael Park",
		"Dr. Priya Sharma", "Dr. Sharma", "Priya Sharma",
		"陈医生", "Park医生", "Sharma医生",
	}
	for _, d := range dentists {
		if strings.Contains(original, d) || strings.Contains(lower, strings.ToLower(d)) {
			return d
		}
	}
	return ""
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
