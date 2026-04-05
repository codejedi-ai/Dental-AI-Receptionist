from livekit.agents import (
    Agent,
    AgentSession,
    AutoSubscribe,
    JobContext,
    RoomInputOptions,
)
from livekit.plugins import noise_cancellation

from app.controller.agent_controller import AgentController
from app.core.observer import AgentEvent
from app.adapters.livekit.brain_adapter import BrainAdapter


class LiveKitController(AgentController):
    def __init__(self, ear, brain, voice, intro_phrase: str):
        super().__init__()
        self.ear = ear
        self.brain = brain
        self.voice = voice
        self.intro_phrase = intro_phrase

    async def run(self, ctx: JobContext):
        self.notify(AgentEvent(type="session_start", payload={"runtime": "livekit"}))
        await ctx.connect(auto_subscribe=AutoSubscribe.AUDIO_ONLY)
        await ctx.wait_for_participant()

        session = AgentSession(
            stt=self.ear.get_stt_plugin(),
            llm=BrainAdapter(self.brain),
            tts=self.voice.get_tts_plugin(),
            vad=ctx.proc.userdata["vad"],
        )

        await session.start(
            room=ctx.room,
            agent=Agent(),
            room_input_options=RoomInputOptions(
                noise_cancellation=noise_cancellation.BVC()
            ),
        )

        if self.intro_phrase:
            await session.say(self.intro_phrase)

        self.notify(AgentEvent(type="session_ready", payload={"runtime": "livekit"}))
