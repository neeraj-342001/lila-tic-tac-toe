import { useCallback, useEffect, useRef, useState } from "react";
import { Client, Session, Socket } from "@heroiclabs/nakama-js";
import { DonePayload, Op, StartPayload, UpdatePayload } from "./protocol";
import "./App.css";

function envHost() {
  return import.meta.env.VITE_NAKAMA_HOST ?? "127.0.0.1";
}
function envPort() {
  return import.meta.env.VITE_NAKAMA_PORT ?? "7350";
}
function envKey() {
  return import.meta.env.VITE_NAKAMA_SERVER_KEY ?? "defaultkey";
}

function envUseSSL(): boolean {
  const v = import.meta.env.VITE_NAKAMA_USE_SSL;
  return v === "true" || v === "1";
}

function deviceId(): string {
  const k = "lila_ttt_device";
  let id = localStorage.getItem(k);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(k, id);
  }
  return id;
}

function markLabel(m: number): string {
  if (m === 1) return "X";
  if (m === 2) return "O";
  return "";
}

export function App() {
  const [status, setStatus] = useState<string>("Disconnected");
  const [error, setError] = useState<string | null>(null);
  const [timedMode, setTimedMode] = useState(false);
  const [session, setSession] = useState<Session | null>(null);
  const [matchId, setMatchId] = useState<string | null>(null);

  const [board, setBoard] = useState<number[]>(() => Array(9).fill(0));
  const [myMark, setMyMark] = useState<number | null>(null);
  const [turnMark, setTurnMark] = useState<number>(1);
  const [deadline, setDeadline] = useState<number | null>(null);
  const [winner, setWinner] = useState<number | null>(null);
  const [line, setLine] = useState<number[] | null>(null);
  const [opponentGone, setOpponentGone] = useState(false);

  const clientRef = useRef<Client | null>(null);
  const socketRef = useRef<Socket | null>(null);
  const matchIdRef = useRef<string | null>(null);

  const connect = useCallback(async () => {
    setError(null);
    setStatus("Connecting…");
    const useSSL = envUseSSL();
    const client = new Client(envKey(), envHost(), envPort(), useSSL);
    clientRef.current = client;
    const sess = await client.authenticateDevice(deviceId(), true);
    setSession(sess);
    const socket = client.createSocket(useSSL, false);
    socketRef.current = socket;

    socket.ondisconnect = () => {
      setStatus("Socket disconnected");
    };

    socket.onmatchpresence = (ev) => {
      if (ev.leaves?.length) setOpponentGone(true);
    };

    socket.onmatchdata = (md) => {
      const op = Number(md.op_code);
      const raw = md.data instanceof Uint8Array ? new TextDecoder().decode(md.data) : String(md.data ?? "");
      try {
        if (op === Op.Start) {
          const p = JSON.parse(raw) as StartPayload;
          setBoard(p.board ?? Array(9).fill(0));
          setTurnMark(p.mark ?? 1);
          setDeadline(p.deadline ?? null);
          setWinner(null);
          setLine(null);
          setOpponentGone(false);
          const uid = sess.user_id;
          setMyMark(uid != null ? (p.marks?.[uid] ?? null) : null);
        } else if (op === Op.Update) {
          const p = JSON.parse(raw) as UpdatePayload;
          setBoard(p.board);
          setTurnMark(p.mark);
          setDeadline(p.deadline ?? null);
        } else if (op === Op.Done) {
          const p = JSON.parse(raw) as DonePayload;
          setBoard(p.board);
          setWinner(p.winner ?? 0);
          setLine(p.winner_positions?.map(Number) ?? null);
          setDeadline(null);
        } else if (op === Op.OpponentLeft) {
          setOpponentGone(true);
        }
      } catch (e) {
        console.warn("match data parse", op, e);
      }
    };

    await socket.connect(sess, true);
    setStatus("Online — find a match");
  }, []);

  useEffect(() => {
    void connect().catch((e: unknown) => {
      setError(e instanceof Error ? e.message : String(e));
      setStatus("Error");
    });
    return () => {
      socketRef.current?.disconnect(true);
      socketRef.current = null;
      clientRef.current = null;
    };
  }, [connect]);

  const findMatch = async () => {
    if (!session || !clientRef.current || !socketRef.current) return;
    setError(null);
    setStatus("Finding match…");
    setWinner(null);
    setLine(null);
    setOpponentGone(false);
    try {
      const res = await clientRef.current.rpc(session, "find_match", { fast: timedMode });
      const body = (res.payload ?? {}) as { match_ids?: string[] };
      const ids = body.match_ids;
      if (!ids?.length) throw new Error("No match id from server");
      const id = ids[0];
      matchIdRef.current = id;
      setMatchId(id);
      await socketRef.current.joinMatch(id);
      setStatus("In match — waiting for opponent / moves");
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
      setStatus("Error");
    }
  };

  const leaveMatch = async () => {
    const sock = socketRef.current;
    const mid = matchIdRef.current;
    if (sock && mid) {
      try {
        await sock.leaveMatch(mid);
      } catch {
        /* ignore */
      }
    }
    matchIdRef.current = null;
    setMatchId(null);
    setBoard(Array(9).fill(0));
    setMyMark(null);
    setWinner(null);
    setLine(null);
    setDeadline(null);
    setStatus("Online — find a match");
  };

  const sendMove = (cell: number) => {
    const sock = socketRef.current;
    const mid = matchIdRef.current;
    if (!sock || !mid || winner != null) return;
    if (myMark == null || turnMark !== myMark) return;
    if (board[cell] !== 0) return;
    const payload = JSON.stringify({ position: cell });
    void sock.sendMatchState(mid, Op.Move, payload);
  };

  const [, setTick] = useState(0);
  useEffect(() => {
    if (deadline == null) return;
    const id = window.setInterval(() => setTick((n) => n + 1), 400);
    return () => window.clearInterval(id);
  }, [deadline]);

  const secondsLeft =
    deadline != null
      ? Math.max(0, Math.floor(deadline - Date.now() / 1000))
      : null;

  return (
    <div className="shell">
      <header className="hero">
        <p className="eyebrow">Lila take-home</p>
        <h1>Tic-Tac-Toe</h1>
        <p className="lede">
          Server-authoritative multiplayer on Nakama. Open two browser windows (or a phone + laptop) to play.
        </p>
      </header>

      <section className="panel">
        <div className="row">
          <span className={`pill ${session ? "ok" : ""}`}>{status}</span>
          {session && (
            <span className="muted mono">
              You: {(session.user_id ?? "").slice(0, 8) || "—"}…
            </span>
          )}
        </div>

        <div className="modes">
          <label className={`mode ${!timedMode ? "active" : ""}`}>
            <input
              type="radio"
              name="mode"
              checked={!timedMode}
              onChange={() => setTimedMode(false)}
            />
            Classic (20s / turn)
          </label>
          <label className={`mode ${timedMode ? "active" : ""}`}>
            <input
              type="radio"
              name="mode"
              checked={timedMode}
              onChange={() => setTimedMode(true)}
            />
            Timed (10s / turn)
          </label>
        </div>

        <div className="actions">
          {!matchId ? (
            <button type="button" className="primary" onClick={() => void findMatch()}>
              Find match
            </button>
          ) : (
            <button type="button" className="ghost" onClick={() => void leaveMatch()}>
              Leave match
            </button>
          )}
        </div>

        {error && <p className="err">{error}</p>}

        {matchId && (
          <div className="game">
            <div className="hud">
              {myMark != null && (
                <p>
                  You are <strong>{markLabel(myMark)}</strong>
                  {deadline != null && secondsLeft != null && (
                    <span className="timer"> · {secondsLeft}s</span>
                  )}
                </p>
              )}
              {winner != null && (
                <p className="result">
                  {winner === 0
                    ? "Draw."
                    : winner === myMark
                      ? "You win."
                      : "You lose."}
                </p>
              )}
              {opponentGone && <p className="warn">Opponent disconnected.</p>}
            </div>

            <div className="grid" role="grid">
              {board.map((cell, i) => (
                <button
                  key={i}
                  type="button"
                  className={`cell ${line?.includes(i) ? "win" : ""}`}
                  disabled={
                    winner != null ||
                    myMark == null ||
                    turnMark !== myMark ||
                    cell !== 0
                  }
                  onClick={() => sendMove(i)}
                >
                  {markLabel(cell)}
                </button>
              ))}
            </div>
          </div>
        )}
      </section>

      <footer className="foot">
        <p>
          Nakama API <span className="mono">{envHost()}:{envPort()}</span> · Configure{" "}
          <code>VITE_NAKAMA_*</code> for production.
        </p>
      </footer>
    </div>
  );
}
