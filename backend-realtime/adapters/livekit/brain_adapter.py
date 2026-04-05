from livekit.agents import llm


class BrainAdapter(llm.LLM):
    def __init__(self, brain):
        super().__init__()
        self.brain = brain

    async def chat(
        self,
        history: llm.ChatContext,
        fnc_ctx: llm.FunctionContext = None,
        temperature: float = None,
        n: int = 1,
        parallel_tool_calls: bool = True,
    ):
        if not history.messages:
            return llm.ChatChunk(choices=[])

        last_msg = history.messages[-1]
        user_text = last_msg.content if last_msg.content else ""

        response_generator = self.brain.think(user_text)
        return BrainStream(response_generator)


class BrainStream:
    def __init__(self, generator):
        self.generator = generator

    async def __aiter__(self):
        async for chunk in self.generator:
            yield llm.ChatChunk(
                choices=[
                    llm.Choice(
                        delta=llm.ChoiceDelta(content=chunk, role="assistant"),
                        index=0,
                    )
                ]
            )
