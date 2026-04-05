from livekit.agents import llm

from app.core.tools import ToolRegistry


def build_function_context(tool_registry: ToolRegistry) -> llm.FunctionContext:
    ctx = llm.FunctionContext()
    for tool in tool_registry.list_tools():
        async def wrapper(_tool=tool, **kwargs):
            return await _tool(**kwargs)

        ctx.register_fnc(
            wrapper,
            name=tool.spec.name,
            description=tool.spec.description,
        )
    return ctx
