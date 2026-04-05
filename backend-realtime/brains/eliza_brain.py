import re
import random
from typing import AsyncGenerator
from app.core.interfaces import Brain

class ElizaBrain(Brain):
    """
    A simple pattern-matching brain based on Weizenbaum's Eliza (1966).
    """
    def __init__(self):
        self.patterns = [
            (r'I need (.*)', [
                "Why do you need {0}?",
                "Would it really help you to get {0}?",
                "Are you sure you need {0}?"
            ]),
            (r'Why don\'?t you ([^\?]*)\??', [
                "Do you really think I don't {0}?",
                "Perhaps eventually I will {0}.",
                "Do you really want me to {0}?"
            ]),
            (r'Why can\'?t I ([^\?]*)\??', [
                "Do you think you should be able to {0}?",
                "If you could {0}, what would you do?",
                "I don't know -- why can't you {0}?"
            ]),
            (r'I am (.*)', [
                "Did you come to me because you are {0}?",
                "How long have you been {0}?",
                "How do you feel about being {0}?"
            ]),
            (r'I\'?m (.*)', [
                "How does being {0} make you feel?",
                "Do you enjoy being {0}?",
                "Why do you tell me you're {0}?"
            ]),
            (r'Are you ([^\?]*)\??', [
                "Why does it matter whether I am {0}?",
                "Would you prefer it if I were not {0}?",
                "Perhaps you believe I am {0}."
            ]),
            (r'What (.*)', [
                "Why do you ask?",
                "How would an answer to that help you?",
                "What do you think?"
            ]),
            (r'How (.*)', [
                "How do you suppose?",
                "Perhaps you can answer your own question.",
                "What is it you're really asking?"
            ]),
            (r'Because (.*)', [
                "Is that the real reason?",
                "What other reasons come to mind?",
                "Does that reason apply to anything else?"
            ]),
            (r'(.*) sorry (.*)', [
                "There are many times when no apology is needed.",
                "What feelings do you have when you apologize?"
            ]),
            (r'Hello(.*)', [
                "Hello... I'm glad you could drop by today.",
                "Hi there... how are you today?",
                "Hello, how are you feeling today?"
            ]),
            (r'(.*)', [
                "Please tell me more.",
                "Let's change focus a bit... Tell me about your family.",
                "Can you elaborate on that?",
                "Why do you say that {0}?",
                "I see.",
                "Very interesting."
            ])
        ]
        self.reflections = {
            "am": "are",
            "was": "were",
            "i": "you",
            "i'd": "you would",
            "i've": "you have",
            "i'll": "you will",
            "my": "your",
            "are": "am",
            "you've": "I have",
            "you'll": "I will",
            "your": "my",
            "yours": "mine",
            "you": "me",
            "me": "you"
        }

    def _reflect(self, fragment):
        tokens = fragment.lower().split()
        for i, token in enumerate(tokens):
            if token in self.reflections:
                tokens[i] = self.reflections[token]
        return ' '.join(tokens)

    async def think(self, user_input: str) -> AsyncGenerator[str, None]:
        # Simple pattern matching
        response = "Tell me more."
        for pattern, responses in self.patterns:
            match = re.match(pattern, user_input, re.IGNORECASE)
            if match:
                response = random.choice(responses)
                if '{0}' in response:
                    phrase = self._reflect(match.group(1))
                    response = response.format(phrase)
                break
        
        # Yield the response (simulating a stream)
        yield response
