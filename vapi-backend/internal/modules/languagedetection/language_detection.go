// Package languagedetection implements Module 1: Language Detection and Initial Response
// for the Dental AI Receptionist system.
//
// It detects whether the caller speaks English or Chinese based on the presence of
// Chinese characters (Unicode \u4e00–\u9fff), then generates an appropriate first response.
package languagedetection

import (
	"log"
	"regexp"
	"strings"
)

// Confidence level for language detection.
type Confidence string

const (
	ConfidenceHigh     Confidence = "high"
	ConfidenceUncertain Confidence = "uncertain"
)

// Language code constants.
const (
	LangZH = "zh"
	LangEN = "en"
)

// Precompiled regex for Chinese character detection.
var chineseCharPattern = regexp.MustCompile(`[\p{Han}]`)

// DetectionResult is the structured output consumed by Module 2.
type DetectionResult struct {
	Module        string     `json:"module"`
	LangCode      string     `json:"lang_code"`
	FirstResponse string     `json:"first_response"`
	Confidence    Confidence `json:"confidence"`
	Metadata      Metadata   `json:"metadata"`
}

// Metadata contains diagnostic info about the detection.
type Metadata struct {
	InputSentence    string `json:"input_sentence"`
	DetectedChinese  bool   `json:"detected_chinese"`
	DetectionMethod  string `json:"detection_method"`
}

// Detect analyzes the user's first sentence and returns a DetectionResult.
//
// Detection rules:
//   - Contains any Chinese characters (\p{Han}) → lang_code = "zh"
//   - English/Latin only → lang_code = "en"
//   - Uncertain or mixed (e.g., pinyin + English) → lang_code = "en", confidence = "uncertain"
func Detect(userFirstSentence string) DetectionResult {
	result := DetectionResult{
		Module: "language_detection",
		Metadata: Metadata{
			InputSentence:   userFirstSentence,
			DetectionMethod: "regex",
		},
	}

	// Edge case: empty or whitespace-only input
	trimmed := strings.TrimSpace(userFirstSentence)
	if trimmed == "" {
		log.Println("[LangDetect] Empty or whitespace-only input")
		result.LangCode = LangEN
		result.Confidence = ConfidenceUncertain
		result.FirstResponse = ENTemplateUnclear
		result.Metadata.DetectedChinese = false
		return result
	}

	// Check for Chinese characters
	hasChinese := chineseCharPattern.MatchString(userFirstSentence)
	result.Metadata.DetectedChinese = hasChinese

	if hasChinese {
		result.LangCode = LangZH
		result.Confidence = ConfidenceHigh
		// Determine which Chinese template to use
		if isIntentStated(trimmed) {
			result.FirstResponse = ZHTemplateIntentStated
		} else {
			result.FirstResponse = ZHTemplateGreeting
		}
		log.Printf("[LangDetect] Detected Chinese: %q → lang=zh", trimmed)
		return result
	}

	// Check for uncertain/mixed: non-ASCII but no Chinese (e.g., pinyin, accented chars)
	hasNonASCII := false
	for _, r := range userFirstSentence {
		if r > 127 {
			hasNonASCII = true
			break
		}
	}

	if hasNonASCII {
		// Mixed/uncertain: default to English but log warning
		result.LangCode = LangEN
		result.Confidence = ConfidenceUncertain
		if isIntentStated(trimmed) {
			result.FirstResponse = ENTemplateIntentStated
		} else {
			result.FirstResponse = ENTemplateUnclear
		}
		log.Printf("[LangDetect] Uncertain input: %q → lang=en (logged for review)", trimmed)
		return result
	}

	// Pure ASCII/Latin — English
	result.LangCode = LangEN
	result.Confidence = ConfidenceHigh
	if isIntentStated(trimmed) {
		result.FirstResponse = ENTemplateIntentStated
	} else if isGreetingOnly(trimmed) {
		result.FirstResponse = ENTemplateGreeting
	} else {
		result.FirstResponse = ENTemplateUnclear
	}
	log.Printf("[LangDetect] Detected English: %q → lang=en", trimmed)
	return result
}

// isIntentStated checks if the input appears to state a clear intent
// (beyond just a greeting).
func isIntentStated(s string) bool {
	lower := strings.ToLower(s)
	intentKeywords := []string{
		"book", "appointment", "schedule", "reserve",
		"cancel", "reschedule", "move", "change",
		"hours", "open", "close",
		"location", "address", "where",
		"service", "root canal", "cleaning", "whitening",
		"emergency", "urgent", "pain", "toothache",
		"预约", "取消", "改期", "时间", "地址", "服务", "急诊",
	}
	for _, kw := range intentKeywords {
		if strings.Contains(lower, kw) || strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// isGreetingOnly checks if the input is primarily a greeting with no other intent.
func isGreetingOnly(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	greetingWords := []string{"hello", "hi", "hey", "good morning", "good afternoon",
		"good evening", "greetings", "nihao", "你好", "喂"}
	for _, gw := range greetingWords {
		if lower == gw || strings.HasPrefix(lower, gw+",") || strings.HasPrefix(lower, gw+"!") {
			return true
		}
	}
	return false
}

// ─── Response Templates ───────────────────────────────────────────

// English response templates (max 2 sentences, warm/polite/professional).
const (
	ENTemplateGreeting       = "Hello! Welcome to Smile Dental Clinic. How can I help you today?"
	ENTemplateIntentStated   = "Thank you for calling Smile Dental. I'd be happy to help you with that."
	ENTemplateUnclear        = "Hello! This is Riley from Smile Dental. Could you tell me how I can assist you?"
)

// Chinese response templates (max 2 sentences, warm/polite/professional).
const (
	ZHTemplateGreeting       = "您好！欢迎来到微笑牙科。请问今天有什么可以帮您的？"
	ZHTemplateIntentStated   = "感谢您致电微笑牙科。我很乐意为您安排。"
	ZHTemplateUnclear        = "您好！我是微笑牙科的Riley。请问您需要什么帮助呢？"
)
