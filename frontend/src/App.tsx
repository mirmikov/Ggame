import { useEffect, useMemo, useState } from "react";
import { api, roomSocket } from "./api/client";
import type { Player, Question, Room, Session, TeamName, TerminalTask } from "./types";

type View = "start" | "create" | "join" | "lobby" | "game" | "results";
const saved = localStorage.getItem("prometheus-session");
const initialSession: Session = saved ? JSON.parse(saved) : { playerId: "", nickname: "", grade: 9 };

export default function App() {
  const [session, setSession] = useState<Session>(initialSession);
  const [view, setView] = useState<View>("start");
  const [room, setRoom] = useState<Room | null>(null);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [tasks, setTasks] = useState<TerminalTask[]>([]);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const updateSession = (next: Session) => {
    setSession(next);
    localStorage.setItem("prometheus-session", JSON.stringify(next));
  };
  useEffect(() => { api.questions(session.grade).then(setQuestions).catch(() => undefined); }, [session.grade]);
  useEffect(() => { api.tasks().then(setTasks).catch(() => undefined); }, []);
  useEffect(() => {
    if (!room?.uniqueServerId) return;
    const socket = roomSocket(room.uniqueServerId);
    socket.onmessage = (message) => {
      const event = JSON.parse(message.data);
      const next: Room = event.payload?.room || event.payload;
      if (next?.uniqueServerId) {
        setRoom(next);
        setView(next.status === "waiting" ? "lobby" : next.status === "finished" ? "results" : "game");
      }
      if (event.payload?.explanation) setNotice(event.payload.explanation);
    };
    return () => socket.close();
  }, [room?.uniqueServerId]);

  const run = async (action: () => Promise<void>) => {
    setError("");
    try { await action(); } catch (e) { setError(e instanceof Error ? e.message : "Неизвестная ошибка"); }
  };

  return <div className="app-shell">
    <div className="scanlines" />
    <header><span className="brand-mark">P//B</span><div><b>PROMETHEUS BATTLE</b><small>ACADEMY NETWORK // ONLINE</small></div><span className="status-dot">SYSTEM READY</span></header>
    {error && <div className="toast error">{error}<button onClick={() => setError("")}>×</button></div>}
    {notice && <div className="toast">{notice}<button onClick={() => setNotice("")}>×</button></div>}
    <main>
      {view === "start" && <Start session={session} setSession={updateSession} onCreate={() => setView("create")} onJoin={() => setView("join")} />}
      {view === "create" && <Create session={session} back={() => setView("start")} create={(data) => run(async () => {
        const result = await api.createRoom(data); updateSession({...session, playerId: result.player.id, roomId: result.room.uniqueServerId}); setRoom(result.room); setView("lobby");
      })} />}
      {view === "join" && <Join back={() => setView("start")} join={(id) => run(async () => {
        const result = await api.joinRoom(id, session.nickname, session.grade); updateSession({...session, playerId: result.player.id, roomId: result.room.uniqueServerId}); setRoom(result.room); setView("lobby");
      })} />}
      {view === "lobby" && room && <Lobby room={room} playerId={session.playerId} choose={(team) => run(async () => setRoom(await api.chooseTeam(room.uniqueServerId, session.playerId, team)))} addBot={() => run(async () => setRoom(await api.addBot(room.uniqueServerId, session.playerId)))} start={() => run(async () => setRoom(await api.start(room.uniqueServerId, session.playerId)))} />}
      {view === "game" && room && <Game room={room} playerId={session.playerId} questions={questions} tasks={tasks} submitTask={(task, value) => run(async () => {
        const result = await api.submitTask(room.uniqueServerId, session.playerId, task, value); setNotice(result.correct ? "DATA MODULE ПОЛУЧЕН // +250" : "OUTPUT НЕ СОВПАЛ"); setRoom(result.room);
      })} answer={(q, a) => run(async () => {
        const result = await api.answer(room.uniqueServerId, session.playerId, q, a); setNotice(`${result.correct ? "ВЕРНО" : "ОШИБКА"} // ${result.explanation}`); setRoom(result.room);
      })} />}
      {view === "results" && room && <Results room={room} restart={() => {setRoom(null); setView("start");}} />}
    </main>
  </div>;
}

function Start({session, setSession, onCreate, onJoin}: {session: Session; setSession: (s: Session) => void; onCreate: () => void; onJoin: () => void}) {
  const ready = session.nickname.trim().length >= 2;
  return <section className="hero grid-bg">
    <div className="eyebrow">GLOBAL NEURAL NETWORK // ACCESS GATE</div>
    <h1>КОД РЕШАЕТ.<br/><span>БАШНИ ПАДАЮТ.</span></h1>
    <p>Отвечай на вопросы, усиливай боевой алгоритм и захвати ядро «Прометея» раньше соперников.</p>
    <div className="panel access-panel">
      <label>ПОЗЫВНОЙ<input value={session.nickname} maxLength={18} placeholder="Введите nickname" onChange={e => setSession({...session, nickname: e.target.value})}/></label>
      <label>КЛАСС<select value={session.grade} onChange={e => setSession({...session, grade: +e.target.value})}><option>9</option><option>10</option><option>11</option></select></label>
      <div className="actions"><button className="primary" disabled={!ready} onClick={onCreate}>СОЗДАТЬ СЕРВЕР</button><button disabled={!ready} onClick={onJoin}>ВОЙТИ ПО ID</button></div>
    </div>
    <div className="factions"><span><i className="blue"/>NEXGEN // PRECISION</span><span><i className="pink"/>OMNISOFT // DISRUPTION</span></div>
  </section>;
}

function Create({session, create, back}: {session: Session; create: (d: unknown) => void; back: () => void}) {
  const [form, setForm] = useState({serverName: "Штурм Прометея", maxPlayers: 6, gradeMode: String(session.grade), gameMode: session.grade === 9 ? "theory_only" : "code_and_theory", settings: {roundDurationSeconds: 600, towerHp: 160, teamPlayerLimit: 3}});
  return <Screen title="CONFIGURE // НОВЫЙ СЕРВЕР" back={back}><div className="panel form-grid">
    <label>НАЗВАНИЕ<input value={form.serverName} onChange={e => setForm({...form, serverName: e.target.value})}/></label>
    <label>ИГРОКОВ<select value={form.maxPlayers} onChange={e => setForm({...form, maxPlayers: +e.target.value})}><option>2</option><option>4</option><option>6</option><option>8</option><option>10</option><option>12</option></select></label>
    <label>КЛАСС<select value={form.gradeMode} onChange={e => setForm({...form, gradeMode: e.target.value})}><option value="9">9</option><option value="10">10</option><option value="11">11</option><option value="mixed">mixed</option></select></label>
    <label>РЕЖИМ<select value={form.gameMode} onChange={e => setForm({...form, gameMode: e.target.value})}><option value="theory_only">Только теория</option><option value="code_and_theory">Код + теория</option><option value="final_pvp">Штурм Башен</option></select></label>
    <label>ДЛИТЕЛЬНОСТЬ, СЕК<input type="number" value={form.settings.roundDurationSeconds} onChange={e => setForm({...form, settings:{...form.settings, roundDurationSeconds:+e.target.value}})}/></label>
    <label>HP БАШНИ<input type="number" value={form.settings.towerHp} onChange={e => setForm({...form, settings:{...form.settings, towerHp:+e.target.value}})}/></label>
    <button className="primary wide" onClick={() => create({...form, nickname: session.nickname, grade: session.grade})}>РАЗВЕРНУТЬ СЕРВЕР</button>
  </div></Screen>;
}

function Join({join, back}: {join: (id: string) => void; back: () => void}) {
  const [id, setId] = useState("");
  return <Screen title="CONNECT // ВХОД НА СЕРВЕР" back={back}><div className="panel join-card"><span className="terminal-line">root@prometheus: awaiting_server_key</span><input className="server-code-input" value={id} placeholder="CYB-XXXXX" onChange={e => setId(e.target.value.toUpperCase())}/><button className="primary" onClick={() => join(id)}>ПОДКЛЮЧИТЬСЯ</button></div></Screen>;
}

function Lobby({room, playerId, choose, addBot, start}: {room: Room; playerId: string; choose: (t: TeamName) => void; addBot: () => void; start: () => void}) {
  const me = room.players[playerId];
  const hasBot = Object.values(room.players).some(p => p.isBot);
  return <Screen title={`LOBBY // ${room.serverName}`} badge={room.uniqueServerId}>
    <div className="lobby-grid"><div className="team-column nex"><h2>NEXGEN</h2><TeamPlayers players={room.players} team="NexGen"/><button onClick={() => choose("NexGen")}>ВСТУПИТЬ В NEXGEN</button></div>
    <div className="panel room-core"><div className="server-id"><small>SERVER ACCESS ID</small><strong>{room.uniqueServerId}</strong></div><div className="setting-list"><span>РЕЖИМ <b>{room.gameMode}</b></span><span>КЛАСС <b>{room.gradeMode}</b></span><span>БАШНЯ <b>{room.settings.towerHp} HP</b></span><span>ИГРОКИ <b>{Object.keys(room.players).length}/{room.maxPlayers}</b></span></div>{me?.isHost ? <><button className="bot-button" disabled={hasBot} onClick={addBot}>{hasBot ? "BOT-7 ПОДКЛЮЧЕН" : "+ ДОБАВИТЬ TEST BOT"}</button><button className="primary pulse" onClick={start}>ЗАПУСТИТЬ БОЙ</button></> : <p className="waiting">Ожидание запуска host...</p>}</div>
    <div className="team-column omni"><h2>OMNISOFT</h2><TeamPlayers players={room.players} team="OmniSoft"/><button onClick={() => choose("OmniSoft")}>ВСТУПИТЬ В OMNISOFT</button></div></div>
  </Screen>;
}

function TeamPlayers({players, team}: {players: Record<string, Player>; team: TeamName}) {
  return <div className="player-list">{Object.values(players).filter(p => p.team === team).map(p => <div className={`player-chip ${p.isBot ? "bot-chip" : ""}`} key={p.id}><span>{p.isBot ? "AI" : p.nickname[0].toUpperCase()}</span>{p.nickname}{p.isBot && <em>TEST BOT</em>}<small>LVL {p.level}</small></div>)}{!Object.values(players).some(p => p.team === team) && <p>СЛОТЫ СВОБОДНЫ</p>}</div>;
}

function laneY(lane: number, laneCount: number) {
  return laneCount <= 1 ? 50 : 16 + lane * 68 / (laneCount - 1);
}
function routePoint(lane: number, position: number, laneCount: number) {
  const y = lane < 0 ? 50 : laneY(lane, laneCount);
  if (position <= 18) return {x: position, y: 50 + (y - 50) * Math.max(0, position - 4) / 14};
  if (position >= 82) return {x: position, y: y + (50 - y) * Math.min(1, (position - 82) / 14)};
  return {x: position, y};
}

function Game({room, playerId, questions, tasks, answer, submitTask}: {room: Room; playerId: string; questions: Question[]; tasks: TerminalTask[]; answer: (q: string, a: number) => void; submitTask: (task: string, value: string) => void}) {
  const me = room.players[playerId];
  const question = questions.find(q => q.id === me?.questionId);
  const [tab, setTab] = useState<"theory"|"terminal">("theory");
  const [taskAnswer, setTaskAnswer] = useState<Record<string,string>>({});
  const laneCount = Math.max(1, ...Object.values(room.units).map(u => u.lane + 1));
  const remaining = Math.max(0, (room.endsAt || 0) - Math.floor(Date.now()/1000));
  const locked = (me?.lockedUntil || 0) > Date.now()/1000;
  return <section className="game-screen">
    <div className="game-top"><Tower team={room.teams.NexGen}/><div className="timer">{Math.floor(remaining/60).toString().padStart(2,"0")}:{(remaining%60).toString().padStart(2,"0")}<small>NETWORK COLLAPSE</small></div><Tower team={room.teams.OmniSoft}/></div>
    <div className="story">{room.storyMessage}</div>
    <div className="arena topdown"><svg className="route-map" viewBox="0 0 100 100" preserveAspectRatio="none">{Array.from({length:laneCount},(_,lane)=>{const y=laneY(lane,laneCount);return <g key={lane}><polyline points={`4,50 18,${y} 82,${y} 96,50`}/>{[28,50,72].map(x=><circle key={`${lane}-${x}`} cx={x} cy={y} r="0.7"/>)}</g>})}</svg><div className="tower-icon left"><img src="/assets/openmoji/tower.svg"/><b>NEX</b></div><div className="tower-icon right"><img src="/assets/openmoji/tower.svg"/><b>OMNI</b></div>{Object.values(room.units).map(u => {const point=routePoint(u.lane,u.position,laneCount);return <div key={u.ownerPlayerId} className={`unit ${u.team === "NexGen" ? "nex" : "omni"} ${u.hp <= 0 ? "down" : ""}`} style={{left:`${point.x}%`,top:`${point.y}%`}}><b>{u.nickname} // P{u.lane+1}</b><img src="/assets/openmoji/robot.svg"/><span style={{width:`${Math.max(0,u.hp/u.maxHp*100)}%`}}/><small>LVL {u.level} // {u.hp}HP</small></div>})}{(room.projectiles || []).map(p => {const from=routePoint(p.fromLane,p.from,laneCount),to=routePoint(p.toLane,p.to,laneCount),angle=Math.atan2(to.y-from.y,to.x-from.x)*180/Math.PI;return <div key={p.id} className={`shot ${p.team === "NexGen" ? "nex" : "omni"}`} style={{left:`${from.x}%`,top:`${from.y}%`,"--shot-to-x":`${to.x}%`,"--shot-to-y":`${to.y}%`,"--shot-angle":`${angle+90}deg`} as React.CSSProperties}><img src="/assets/openmoji/rocket.svg"/></div>})}{(room.projectiles || []).map(p => {const point=routePoint(p.toLane,p.to,laneCount);return <div key={`hit-${p.id}`} className="hit" style={{left:`${point.x}%`,top:`${point.y}%`}}><img src="/assets/openmoji/explosion.svg"/><b>-{p.damage}</b></div>})}</div>
    <div className="combat-grid"><div className="panel stats"><h3>UNIT // {me?.nickname}</h3><Stat name="LEVEL" value={me?.level}/><Stat name="ATTACK" value={me?.attack}/><Stat name="DEFENSE" value={me?.defense}/><Stat name="SPEED" value={me?.speed.toFixed(1)}/><Stat name="SCORE" value={me?.score}/></div>
    <div className="panel question"><div className="tabs"><button className={tab==="theory"?"active":""} onClick={()=>setTab("theory")}>THEORY</button>{room.gameMode !== "theory_only" && <button className={tab==="terminal"?"active":""} onClick={()=>setTab("terminal")}>TERMINAL</button>}</div>{tab === "terminal" ? <div className="tasks">{tasks.map(t => <div key={t.id} className={me?.solvedTasks?.includes(t.id) ? "solved" : ""}><b>{t.title} // +{t.reward}</b><p>{t.description}</p><input value={taskAnswer[t.id]||""} disabled={me?.solvedTasks?.includes(t.id)} placeholder="expected output" onChange={e=>setTaskAnswer({...taskAnswer,[t.id]:e.target.value})}/><button disabled={me?.solvedTasks?.includes(t.id)} onClick={()=>submitTask(t.id,taskAnswer[t.id]||"")}>{me?.solvedTasks?.includes(t.id)?"SOLVED":"EXECUTE"}</button></div>)}</div> : <><div className="question-head"><span>{question?.topic || "SYNC"}</span><b>DIFFICULTY {question?.difficulty || 1}</b></div>{locked ? <div className="locked"><b>THEORY LOCKED</b><span>Слишком много ошибок. Перезагрузка модуля...</span></div> : question ? <><h2>{question.text}</h2><div className="options">{question.options.map((o,i) => <button key={o} onClick={() => answer(question.id,i)}><kbd>{String.fromCharCode(65+i)}</kbd>{o}</button>)}</div></> : <p>Синхронизация вопроса...</p>}</>}</div>
    <div className="panel score"><h3>TEAM FEED</h3><div className="big-score blue-text">{room.teams.NexGen.score}</div><small>NEXGEN DATA</small><div className="big-score pink-text">{room.teams.OmniSoft.score}</div><small>OMNISOFT DATA</small><hr/><span>Верно: {me?.correctAnswers}</span><span>Ошибки: {me?.wrongAnswers}</span></div></div>
  </section>;
}

function Tower({team}: {team: {name: TeamName; towerHp: number; maxTowerHp: number}}) {
  const pct = Math.max(0, team.towerHp/team.maxTowerHp*100);
  return <div className={`tower-hp ${team.name === "NexGen" ? "nex" : "omni"}`}><div><b>{team.name} TOWER</b><span>{Math.max(0,team.towerHp)} / {team.maxTowerHp}</span></div><i><em style={{width:`${pct}%`}}/></i></div>;
}
function Stat({name,value}: {name:string; value:string|number|undefined}) { return <div className="stat"><span>{name}</span><b>{value}</b></div>; }

function Results({room, restart}: {room: Room; restart: () => void}) {
  const players = Object.values(room.players).sort((a,b) => b.score-a.score);
  return <Screen title="MATCH COMPLETE // РЕЗУЛЬТАТЫ"><div className={`winner ${room.winner === "NexGen" ? "blue-text" : "pink-text"}`}><small>ПОБЕДИТЕЛЬ</small>{room.winner}</div><p className="result-story">{room.storyMessage}</p><div className="panel leaderboard">{players.map((p,i) => <div key={p.id}><b>#{i+1}</b><span>{p.nickname}<small>{p.team}</small></span><span>{p.correctAnswers} верно</span><span>{p.wrongAnswers} ошибок</span><strong>{p.score}</strong></div>)}</div><button className="primary center" onClick={restart}>ВЕРНУТЬСЯ В СЕТЬ</button></Screen>;
}

function Screen({title, badge, back, children}: {title: string; badge?: string; back?: () => void; children: React.ReactNode}) {
  return <section className="screen"><div className="screen-title">{back && <button onClick={back}>← НАЗАД</button>}<h1>{title}</h1>{badge && <strong>{badge}</strong>}</div>{children}</section>;
}
