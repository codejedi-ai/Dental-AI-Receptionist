from livekit.plugins import rime
from app.core.interfaces import Voice

class RimeVoice(Voice):
    def __init__(self, speaker: str = "celeste", model: str = "arcana", speed: float = 1.0):
        self.speaker = speaker
        self.model = model
        self.speed = speed

    def get_tts_plugin(self):
        return rime.TTS(
            model=self.model,
            speaker=self.speaker,
            speed_alpha=self.speed
        )
