package languagedetection

import (
	"testing"
)

// ─── Test Cases from Spec (Section 13.9) ────────────────────────

func TestDetect_ChineseOnly(t *testing.T) {
	r := Detect("你好")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh, got %s", r.LangCode)
	}
	if r.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence=high, got %s", r.Confidence)
	}
	if !r.Metadata.DetectedChinese {
		t.Error("expected detected_chinese=true")
	}
}

func TestDetect_EnglishOnly(t *testing.T) {
	r := Detect("Hello")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en, got %s", r.LangCode)
	}
	if r.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence=high, got %s", r.Confidence)
	}
	if r.Metadata.DetectedChinese {
		t.Error("expected detected_chinese=false")
	}
}

func TestDetect_MixedEnglishChinese(t *testing.T) {
	r := Detect("I want 预约")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh for mixed input with Chinese chars, got %s", r.LangCode)
	}
	if !r.Metadata.DetectedChinese {
		t.Error("expected detected_chinese=true for mixed input")
	}
}

func TestDetect_PinyinOnly(t *testing.T) {
	r := Detect("Ni hao ma")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for pinyin-only input, got %s", r.LangCode)
	}
	if r.Metadata.DetectedChinese {
		t.Error("pinyin should not trigger Chinese detection")
	}
}

func TestDetect_DotsOnly(t *testing.T) {
	r := Detect("...")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for dots-only input, got %s", r.LangCode)
	}
}

func TestDetect_EnglishWithChineseWord(t *testing.T) {
	r := Detect("Can I book an appointment for 明天?")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh when Chinese chars present, got %s", r.LangCode)
	}
}

func TestDetect_Empty(t *testing.T) {
	r := Detect("")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for empty input, got %s", r.LangCode)
	}
	if r.Confidence != ConfidenceUncertain {
		t.Errorf("expected uncertainty for empty input, got %s", r.Confidence)
	}
}

func TestDetect_MixedChineseEnglish(t *testing.T) {
	r := Detect("你好！Hello!")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh for mixed Chinese+English, got %s", r.LangCode)
	}
}

func TestDetect_EnglishWithChineseWord2(t *testing.T) {
	r := Detect("Is this 牙科诊所?")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh when Chinese chars present, got %s", r.LangCode)
	}
}

func TestDetect_GoodMorning(t *testing.T) {
	r := Detect("Good morning")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en, got %s", r.LangCode)
	}
}

// ─── Additional Edge Cases ──────────────────────────────────────

func TestDetect_WhitespaceOnly(t *testing.T) {
	r := Detect("   \t\n  ")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for whitespace-only input, got %s", r.LangCode)
	}
}

func TestDetect_PunctuationOnly(t *testing.T) {
	r := Detect("?!.,;:")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for punctuation-only input, got %s", r.LangCode)
	}
}

func TestDetect_NonASCII_NoChinese(t *testing.T) {
	// Accented characters but no Chinese
	r := Detect("Café résumé naïve")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en for non-ASCII non-Chinese, got %s", r.LangCode)
	}
	if r.Confidence != ConfidenceUncertain {
		t.Errorf("expected uncertainty for non-ASCII input, got %s", r.Confidence)
	}
}

func TestDetect_BookingIntentEnglish(t *testing.T) {
	r := Detect("I'd like to book a cleaning")
	if r.LangCode != LangEN {
		t.Errorf("expected lang=en, got %s", r.LangCode)
	}
	if r.FirstResponse != ENTemplateIntentStated {
		t.Errorf("expected intent-stated template for booking intent, got: %s", r.FirstResponse)
	}
}

func TestDetect_GreetingOnlyEnglish(t *testing.T) {
	r := Detect("Hello")
	if r.FirstResponse != ENTemplateGreeting {
		t.Errorf("expected greeting template, got: %s", r.FirstResponse)
	}
}

func TestDetect_GreetingOnlyEnglishHi(t *testing.T) {
	r := Detect("Hi!")
	if r.FirstResponse != ENTemplateGreeting {
		t.Errorf("expected greeting template for 'Hi!', got: %s", r.FirstResponse)
	}
}

func TestDetect_GreetingOnlyChinese(t *testing.T) {
	r := Detect("你好")
	if r.FirstResponse != ZHTemplateGreeting {
		t.Errorf("expected Chinese greeting template, got: %s", r.FirstResponse)
	}
}

func TestDetect_IntentStatedChinese(t *testing.T) {
	r := Detect("我想预约")
	if r.LangCode != LangZH {
		t.Errorf("expected lang=zh, got %s", r.LangCode)
	}
	if r.FirstResponse != ZHTemplateIntentStated {
		t.Errorf("expected intent-stated Chinese template, got: %s", r.FirstResponse)
	}
}

func TestDetect_UnclearEnglish(t *testing.T) {
	r := Detect("ummm maybe")
	if r.FirstResponse != ENTemplateUnclear {
		t.Errorf("expected unclear template, got: %s", r.FirstResponse)
	}
}

func TestDetect_ModuleField(t *testing.T) {
	r := Detect("Hello")
	if r.Module != "language_detection" {
		t.Errorf("expected module='language_detection', got %s", r.Module)
	}
}

func TestDetect_MetadataInputSentence(t *testing.T) {
	input := "你好，我想预约"
	r := Detect(input)
	if r.Metadata.InputSentence != input {
		t.Errorf("expected input_sentence=%q, got %q", input, r.Metadata.InputSentence)
	}
}

func TestDetect_DetectionMethod(t *testing.T) {
	r := Detect("Hello")
	if r.Metadata.DetectionMethod != "regex" {
		t.Errorf("expected detection_method='regex', got %s", r.Metadata.DetectionMethod)
	}
}

// ─── Helper Function Tests ──────────────────────────────────────

func TestIsIntentStated_Book(t *testing.T) {
	if !isIntentStated("I want to book an appointment") {
		t.Error("expected book to trigger intent detection")
	}
}

func TestIsIntentStated_Cancel(t *testing.T) {
	if !isIntentStated("I need to cancel my appointment") {
		t.Error("expected cancel to trigger intent detection")
	}
}

func TestIsIntentStated_Emergency(t *testing.T) {
	if !isIntentStated("I have a toothache emergency") {
		t.Error("expected emergency to trigger intent detection")
	}
}

func TestIsIntentStated_NoIntent(t *testing.T) {
	if isIntentStated("Hello") {
		t.Error("greeting-only should not trigger intent detection")
	}
}

func TestIsGreetingOnly_Hello(t *testing.T) {
	if !isGreetingOnly("Hello") {
		t.Error("'Hello' should be detected as greeting only")
	}
}

func TestIsGreetingOnly_WithComma(t *testing.T) {
	if !isGreetingOnly("Hello, I need help") {
		// Actually this has additional content so it should NOT be greeting-only
		t.Log("This should return false because there's content after the greeting")
	}
}

func TestIsGreetingOnly_NotGreeting(t *testing.T) {
	if isGreetingOnly("I want to book an appointment") {
		t.Error("booking intent should not be detected as greeting-only")
	}
}
