import type { WSMessage } from "./types";

export interface WSHandlers {
  onState?: (msg: Extract<WSMessage, { type: "state" }>) => void;
  onError?: (msg: Extract<WSMessage, { type: "error" }>) => void;
  onGameOver?: (msg: Extract<WSMessage, { type: "gameOver" }>) => void;
  onClose?: () => void;
}

export interface WSController {
  send: (msg: object) => void;
  close: () => void;
}

function wsBase(): string {
  const api = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
  return api.replace(/^http/, "ws");
}

export function connectGameSocket(
  gameId: string,
  token: string,
  handlers: WSHandlers
): WSController {
  let ws: WebSocket;
  let closed = false;
  let reconnected = false;

  function connect() {
    ws = new WebSocket(`${wsBase()}/ws`);

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: "join", gameId, token }));
    };

    ws.onmessage = (ev) => {
      let msg: WSMessage;
      try {
        msg = JSON.parse(ev.data as string);
      } catch {
        return;
      }
      if (msg.type === "state") handlers.onState?.(msg);
      else if (msg.type === "error") handlers.onError?.(msg);
      else if (msg.type === "gameOver") handlers.onGameOver?.(msg);
    };

    ws.onclose = (ev) => {
      if (closed) {
        handlers.onClose?.();
        return;
      }
      // Single auto-reconnect on unexpected close
      if (!reconnected && !ev.wasClean) {
        reconnected = true;
        setTimeout(connect, 1500);
      } else {
        handlers.onClose?.();
      }
    };

    ws.onerror = () => {
      // errors always followed by close
    };
  }

  connect();

  return {
    send: (msg: object) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(msg));
      }
    },
    close: () => {
      closed = true;
      ws.close();
    },
  };
}
