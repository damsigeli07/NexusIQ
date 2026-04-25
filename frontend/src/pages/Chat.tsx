import { useEffect, useRef, useState } from "react";
import { getToken, getTenant } from "../store/auth";
import { Send, Bot, User } from "lucide-react";

type Message = { role: "user" | "assistant"; text: string; sources?: string };

export default function Chat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input,    setInput]    = useState("");
  const [streaming,setStreaming]= useState(false);
  const wsRef   = useRef<WebSocket | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const send = () => {
    if (!input.trim() || streaming) return;

    const question = input.trim();
    setInput("");
    setMessages(m => [...m, { role: "user", text: question }]);
    setStreaming(true);

    const token    = getToken();
    const tenantId = getTenant();
    const ws = new WebSocket(
      `ws://localhost:8080/api/chat?token=${token}&tenant_id=${tenantId}`
    );
    wsRef.current = ws;

    // Add empty assistant message we'll stream into
    setMessages(m => [...m, { role: "assistant", text: "" }]);

    ws.onopen = () => {
      ws.send(JSON.stringify({ question }));
    };

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);

      if (msg.type === "token" || msg.type === "cached") {
        setMessages(m => {
          const updated = [...m];
          updated[updated.length - 1] = {
            ...updated[updated.length - 1],
            text: updated[updated.length - 1].text + msg.content,
          };
          return updated;
        });
      }

      if (msg.type === "sources") {
        setMessages(m => {
          const updated = [...m];
          updated[updated.length - 1] = {
            ...updated[updated.length - 1],
            sources: msg.content,
          };
          return updated;
        });
      }

      if (msg.type === "done" || msg.type === "error") {
        setStreaming(false);
        ws.close();
      }
    };

    ws.onerror = () => {
      setStreaming(false);
      setMessages(m => {
        const updated = [...m];
        updated[updated.length - 1].text = "Connection error — is the server running?";
        return updated;
      });
    };
  };

  return (
    <div className="h-full flex flex-col">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        {messages.length === 0 && (
          <div className="h-full flex flex-col items-center justify-center text-gray-500 gap-2">
            <Bot size={40} strokeWidth={1} />
            <p className="text-sm">Ask anything about your uploaded documents</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex gap-3 ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            {msg.role === "assistant" && (
              <div className="w-7 h-7 rounded-full bg-indigo-600 flex items-center justify-center shrink-0 mt-1">
                <Bot size={14} />
              </div>
            )}
            <div className={`max-w-[75%] rounded-xl px-4 py-3 text-sm leading-relaxed ${
              msg.role === "user"
                ? "bg-indigo-600 text-white"
                : "bg-gray-800 text-gray-100"
            }`}>
              {msg.text || <span className="animate-pulse text-gray-400">▋</span>}
              {msg.sources && (
                <p className="text-xs text-gray-400 mt-2 border-t border-gray-700 pt-2">
                  Sources: {msg.sources}
                </p>
              )}
            </div>
            {msg.role === "user" && (
              <div className="w-7 h-7 rounded-full bg-gray-700 flex items-center justify-center shrink-0 mt-1">
                <User size={14} />
              </div>
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t border-gray-800">
        <div className="flex gap-2">
          <input
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === "Enter" && send()}
            placeholder="Ask a question about your documents…"
            className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-indigo-500"
            disabled={streaming}
          />
          <button
            onClick={send} disabled={streaming || !input.trim()}
            className="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 px-4 py-2 rounded-lg transition-colors">
            <Send size={16} />
          </button>
        </div>
      </div>
    </div>
  );
}