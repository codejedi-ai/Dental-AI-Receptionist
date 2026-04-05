from livekit.plugins import openai
from app.core.interfaces import Ear

class OpenAIEar(Ear):
    def get_stt_plugin(self):
        return openai.STT()
