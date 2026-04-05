"use client";

import { useEffect, useRef, useState } from "react";

interface TranscriptMessage {
  role: "assistant" | "user";
  text: string;
}

export default function VoiceWidget() {
  const [isActive, setIsActive] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [transcript, setTranscript] = useState<TranscriptMessage[]>([]);
  const vapiRef = useRef<any>(null);
  const transcriptEndRef = useRef<HTMLDivElement>(null);

  const publicKey = process.env.NEXT_PUBLIC_VAPI_PUBLIC_KEY;
  const assistantId = process.env.NEXT_PUBLIC_VAPI_ASSISTANT_ID;

  useEffect(() => {
    transcriptEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [transcript]);

  const startCall = async () => {
    if (!publicKey || !assistantId) {
      setTranscript([
        { role: "assistant", text: "Voice assistant is not configured. Please set VAPI keys in environment variables." },
      ]);
      setIsOpen(true);
      return;
    }

    try {
      setIsConnecting(true);
      setIsOpen(true);

      const { default: Vapi } = await import("@vapi-ai/web");
      const vapi = new Vapi(publicKey);
      vapiRef.current = vapi;

      vapi.on("call-start", () => {
        setIsActive(true);
        setIsConnecting(false);
        setTranscript([{ role: "assistant", text: "Hi! I'm Lisa, your AI receptionist at Smile Dental. How can I help you today?" }]);
      });

      vapi.on("call-end", () => {
        setIsActive(false);
        setIsConnecting(false);
      });

      vapi.on("message", (message: any) => {
        if (message.type === "transcript" && message.transcriptType === "final") {
          setTranscript((prev) => [...prev, { role: message.role, text: message.transcript }]);
        }
      });

      vapi.on("error", (error: any) => {
        console.error("Vapi error:", error);
        setIsActive(false);
        setIsConnecting(false);
        setTranscript((prev) => [...prev, { role: "assistant", text: "Sorry, there was a connection error. Please try again." }]);
      });

      await vapi.start(assistantId);
    } catch (error) {
      console.error("Failed to start call:", error);
      setIsConnecting(false);
      setTranscript((prev) => [...prev, { role: "assistant", text: "Failed to start voice call. Please try again later." }]);
    }
  };

  const endCall = () => {
    if (vapiRef.current) {
      vapiRef.current.stop();
      vapiRef.current = null;
    }
    setIsActive(false);
    setIsConnecting(false);
  };

  return (
    <>
      {/* Chat panel */}
      {isOpen && (
        <div className="fixed bottom-24 right-4 sm:right-6 w-80 sm:w-96 bg-white rounded-2xl shadow-2xl border border-gray-200 z-50 overflow-hidden">
          <div className="bg-teal-600 text-white p-4 flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-xl">🎙️</span>
              <div>
                <p className="font-semibold text-sm">Lisa — AI Receptionist</p>
                <p className="text-xs text-teal-100">
                  {isActive ? "Listening..." : isConnecting ? "Connecting..." : "Click to start"}
                </p>
              </div>
            </div>
            <button
              onClick={() => {
                setIsOpen(false);
                if (isActive) endCall();
              }}
              className="text-white/80 hover:text-white"
            >
              ✕
            </button>
          </div>

          <div className="h-64 overflow-y-auto p-4 space-y-3 bg-gray-50">
            {transcript.length === 0 && (
              <p className="text-center text-gray-400 text-sm mt-8">
                Click the microphone button to start talking with Lisa
              </p>
            )}
            {transcript.map((msg, i) => (
              <div key={i} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
                <div
                  className={`max-w-[80%] rounded-xl px-3 py-2 text-sm ${
                    msg.role === "user"
                      ? "bg-teal-600 text-white"
                      : "bg-white border border-gray-200 text-gray-700"
                  }`}
                >
                  {msg.text}
                </div>
              </div>
            ))}
            <div ref={transcriptEndRef} />
          </div>

          <div className="p-3 bg-white border-t border-gray-100 flex justify-center">
            {!isActive ? (
              <button
                onClick={startCall}
                disabled={isConnecting}
                className="bg-teal-600 hover:bg-teal-700 disabled:bg-teal-400 text-white px-6 py-2 rounded-full text-sm font-medium transition-colors flex items-center space-x-2"
              >
                <span>🎤</span>
                <span>{isConnecting ? "Connecting..." : "Start Talking"}</span>
              </button>
            ) : (
              <button
                onClick={endCall}
                className="bg-red-500 hover:bg-red-600 text-white px-6 py-2 rounded-full text-sm font-medium transition-colors flex items-center space-x-2"
              >
                <span>⏹️</span>
                <span>End Call</span>
              </button>
            )}
          </div>
        </div>
      )}

      {/* Floating button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`fixed bottom-6 right-4 sm:right-6 z-50 w-14 h-14 rounded-full shadow-lg flex items-center justify-center text-white text-2xl transition-all hover:scale-110 ${
          isActive ? "bg-red-500 animate-pulse" : "bg-teal-600 hover:bg-teal-700"
        }`}
        title="Talk to Lisa, our AI receptionist"
      >
        {isActive ? "🎙️" : "💬"}
      </button>
    </>
  );
}
