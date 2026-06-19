import { useEffect, useMemo, useState } from "react";
import type { CSSProperties, ReactNode } from "react";
import { api, roomSocket } from "./api/client";
import type {
  Player,
  Question,
  Room,
  Session,
  TeamName,
  TerminalTask,
  Unit,
  User,
} from "./types";

type View = "auth" | "start" | "create" | "join" | "lobby" | "game" | "results";

function readSession(): Session {
  try {
    const saved = localStorage.getItem("prometheus-session");
    return saved ? JSON.parse(saved) : { playerId: "", nickname: "", grade: 9 };
  } catch {
    return { playerId: "", nickname: "", grade: 9 };
  }
}

function roomIdFromPath() {
  const match = window.location.pathname.match(/^\/(CYB-[A-Z0-9]+)\/?$/i);
  return match ? match[1].toUpperCase() : "";
}

function roomPath(roomId: string) {
  return `/${roomId.toUpperCase()}`;
}

function showRoomInAddress(roomId: string) {
  const next = roomPath(roomId);
  if (window.location.pathname !== next) {
    window.history.pushState({}, "", next);
  }
}

function showHomeInAddress() {
  if (window.location.pathname !== "/") {
    window.history.pushState({}, "", "/");
  }
}

export default function App() {
  const [session, setSession] = useState<Session>(readSession);
  const [view, setView] = useState<View>("start");
  const [room, setRoom] = useState<Room | null>(null);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [tasks, setTasks] = useState<TerminalTask[]>([]);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [user, setUser] = useState<User | null>(null);
  const [authChecked, setAuthChecked] = useState(false);

  const updateSession = (next: Session) => {
    setSession(next);
    localStorage.setItem("prometheus-session", JSON.stringify(next));
  };

  useEffect(() => {
    api
      .me()
      .then(({ user }) => {
        setUser(user);
        const saved = readSession();
        const nextSession = {
          ...saved,
          nickname: user.displayName,
          grade: user.grade,
        };
        updateSession(nextSession);
        const urlRoomId = roomIdFromPath();
        if (urlRoomId) {
          return api.getRoom(urlRoomId).then((room) => {
            updateSession({ ...nextSession, roomId: room.uniqueServerId });
            if (!nextSession.playerId || !room.players[nextSession.playerId]) {
              setView("join");
              return;
            }
            setRoom(room);
            setView(
              room.status === "waiting"
                ? "lobby"
                : room.status === "finished"
                  ? "results"
                  : "game",
            );
          });
        }
        return undefined;
      })
      .catch((e) => {
        if (roomIdFromPath()) {
          setError(e instanceof Error ? e.message : "Комната не найдена");
        }
      })
      .finally(() => setAuthChecked(true));
  }, []);

  useEffect(() => {
    api
      .questions(session.grade)
      .then(setQuestions)
      .catch(() => undefined);
  }, [session.grade]);

  useEffect(() => {
    api
      .tasks()
      .then(setTasks)
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!room?.uniqueServerId) return;
    const socket = roomSocket(room.uniqueServerId);
    socket.onmessage = (message) => {
      try {
        const event = JSON.parse(message.data);
        const next: Room = event.payload?.room || event.payload;
        if (next?.uniqueServerId) {
          setRoom(next);
          setView(
            next.status === "waiting"
              ? "lobby"
              : next.status === "finished"
                ? "results"
                : "game",
          );
        }
      } catch {
        setError("Получено повреждённое сообщение от сервера");
      }
    };
    socket.onerror = () => setError("Потеряно соединение с сервером");
    return () => socket.close();
  }, [room?.uniqueServerId]);

  const run = async (action: () => Promise<void>) => {
    setError("");
    try {
      await action();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Неизвестная ошибка");
    }
  };

  return (
    <div className="app-shell">
      <div className="scanlines" />
      <header>
        <span className="brand-mark">P//B</span>
        <div>
          <b>PROMETHEUS TOURNAMENT</b>
          <small>QUALIFIER + FINAL // ONLINE</small>
        </div>
        <span className="status-dot">SYSTEM READY</span>
        {user && (
          <button
            className="logout-button"
            onClick={() =>
              run(async () => {
                await api.logout();
                setUser(null);
                setRoom(null);
                showHomeInAddress();
                setView("start");
              })
            }
          >
            ВЫЙТИ
          </button>
        )}
      </header>
      {error && <Toast text={error} error close={() => setError("")} />}
      {notice && <Toast text={notice} close={() => setNotice("")} />}
      <main>
        {!authChecked && (
          <section className="hero grid-bg">
            <div className="eyebrow">AUTH // DATABASE</div>
            <h1>ЗАГРУЗКА</h1>
          </section>
        )}
        {authChecked && !user && (
          <AuthGate
            onAuth={(nextUser) => {
              const urlRoomId = roomIdFromPath();
              setUser(nextUser);
              updateSession({
                playerId: "",
                nickname: nextUser.displayName,
                grade: nextUser.grade,
                roomId: urlRoomId || undefined,
              });
              setView(urlRoomId ? "join" : "start");
            }}
            run={run}
          />
        )}
        {authChecked && user && (
          <>
        {view === "start" && (
          <Start
            session={session}
            setSession={updateSession}
            onCreate={() => setView("create")}
            onJoin={() => setView("join")}
          />
        )}
        {view === "create" && (
          <Create
            session={session}
            back={() => setView("start")}
            create={(data) =>
              run(async () => {
                const result = await api.createRoom(data);
                updateSession({
                  ...session,
                  playerId: result.player.id,
                  roomId: result.room.uniqueServerId,
                  role: "organizer",
                });
                setRoom(result.room);
                showRoomInAddress(result.room.uniqueServerId);
                setView("lobby");
              })
            }
          />
        )}
        {view === "join" && (
          <Join
            back={() => setView("start")}
            join={(id) =>
              run(async () => {
                const result = await api.joinRoom(
                  id,
                  session.nickname,
                  session.grade,
                );
                updateSession({
                  ...session,
                  playerId: result.player.id,
                  roomId: result.room.uniqueServerId,
                  role: "participant",
                });
                setRoom(result.room);
                showRoomInAddress(result.room.uniqueServerId);
                setView("lobby");
              })
            }
          />
        )}
        {view === "lobby" && room && (
          <Lobby
            room={room}
            playerId={session.playerId}
            choose={(team) =>
              run(async () =>
                setRoom(
                  await api.chooseTeam(
                    room.uniqueServerId,
                    session.playerId,
                    team,
                  ),
                ),
              )
            }
            chooseQualifier={(teamId) =>
              run(async () =>
                setRoom(
                  await api.chooseQualifierTeam(
                    room.uniqueServerId,
                    session.playerId,
                    teamId,
                  ),
                ),
              )
            }
            start={() =>
              run(async () =>
                setRoom(await api.start(room.uniqueServerId, session.playerId)),
              )
            }
            finish={() =>
              run(async () => {
                setRoom(await api.finish(room.uniqueServerId, session.playerId));
                setView("results");
              })
            }
          />
        )}
        {view === "game" && room && (
          <TournamentGame
            room={room}
            playerId={session.playerId}
            questions={questions}
            tasks={tasks}
            finish={() =>
              run(async () => {
                setRoom(await api.finish(room.uniqueServerId, session.playerId));
                setView("results");
              })
            }
            submitTask={(task, value) =>
              run(async () => {
                const result = await api.submitTask(
                  room.uniqueServerId,
                  session.playerId,
                  task,
                  value,
                );
                setNotice(
                  result.correct
                    ? room.gameMode === "qualifier"
                      ? "КОД ПРИНЯТ // КОМАНДА ПРОДВИНУЛАСЬ К ЗОНЕ"
                      : "КОД ПРИНЯТ // БОЕВОЙ МОДУЛЬ УСИЛЕН"
                    : "КОД НЕ ПРОШЁЛ ПРОВЕРКУ",
                );
                setRoom(result.room);
              })
            }
            answer={(q, a) =>
              run(async () => {
                const result = await api.answer(
                  room.uniqueServerId,
                  session.playerId,
                  q,
                  a,
                );
                setNotice(
                  `${result.correct ? (room.gameMode === "qualifier" ? "ВЕРНО // ДВИЖЕНИЕ К ЗОНЕ" : "ВЕРНО // ПОЛУЧЕН БАФФ") : "ОШИБКА"} — ${result.explanation}`,
                );
                setRoom(result.room);
              })
            }
          />
        )}
        {view === "results" && room && (
          <Results
            room={room}
            restart={() => {
              setRoom(null);
              showHomeInAddress();
              setView("start");
            }}
          />
        )}
          </>
        )}
      </main>
    </div>
  );
}

function Toast({
  text,
  error,
  close,
}: {
  text: string;
  error?: boolean;
  close: () => void;
}) {
  return (
    <div className={`toast ${error ? "error" : ""}`} role="status">
      {text}
      <button onClick={close}>×</button>
    </div>
  );
}

function AuthGate({
  onAuth,
  run,
}: {
  onAuth: (user: User) => void;
  run: (action: () => Promise<void>) => void;
}) {
  const [mode, setMode] = useState<"login" | "register">("register");
  const [form, setForm] = useState({
    email: "",
    displayName: "",
    password: "",
    grade: 9,
  });
  const isRegister = mode === "register";
  return (
    <section className="hero grid-bg">
      <div className="eyebrow">ACCOUNT // TOURNAMENT DATABASE</div>
      <h1>
        ВХОД В СИСТЕМУ
        <br />
        <span>ТУРНИРЫ СОХРАНЯЮТСЯ.</span>
      </h1>
      <div className="panel auth-panel">
        <div className="auth-tabs">
          <button
            className={isRegister ? "active" : ""}
            onClick={() => setMode("register")}
          >
            РЕГИСТРАЦИЯ
          </button>
          <button
            className={!isRegister ? "active" : ""}
            onClick={() => setMode("login")}
          >
            ВХОД
          </button>
        </div>
        <label>
          EMAIL
          <input
            type="email"
            value={form.email}
            placeholder="name@example.com"
            onChange={(e) => setForm({ ...form, email: e.target.value })}
          />
        </label>
        {isRegister && (
          <>
            <label>
              ИМЯ / ПОЗЫВНОЙ
              <input
                value={form.displayName}
                maxLength={24}
                placeholder="Например, Алекс"
                onChange={(e) =>
                  setForm({ ...form, displayName: e.target.value })
                }
              />
            </label>
            <label>
              КЛАСС
              <select
                value={form.grade}
                onChange={(e) => setForm({ ...form, grade: +e.target.value })}
              >
                <option>9</option>
                <option>10</option>
                <option>11</option>
              </select>
            </label>
          </>
        )}
        <label>
          ПАРОЛЬ
          <input
            type="password"
            value={form.password}
            minLength={6}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
          />
        </label>
        <button
          className="primary wide"
          onClick={() =>
            run(async () => {
              const result = isRegister
                ? await api.register(form)
                : await api.login(form.email, form.password);
              onAuth(result.user);
            })
          }
        >
          {isRegister ? "СОЗДАТЬ АККАУНТ" : "ВОЙТИ"}
        </button>
      </div>
    </section>
  );
}

function Start({
  session,
  setSession,
  onCreate,
  onJoin,
}: {
  session: Session;
  setSession: (s: Session) => void;
  onCreate: () => void;
  onJoin: () => void;
}) {
  const ready = session.nickname.trim().length >= 2;
  return (
    <section className="hero grid-bg">
      <div className="eyebrow">SCHOOL CYBER TOURNAMENT // 9–11 КЛАССЫ</div>
      <h1>
        ДУМАЙ. ПИШИ КОД.
        <br />
        <span>ЗАХВАТЫВАЙ ЗОНУ.</span>
      </h1>
      <p>
        В отборочном туре команды движутся к центральной зоне и вытесняют
        соперников. В финале две стороны сражаются друг с другом.
      </p>
      <div className="panel access-panel">
        <label>
          ИМЯ / ПОЗЫВНОЙ
          <input
            value={session.nickname}
            maxLength={24}
            placeholder="Например, Алекс"
            onChange={(e) =>
              setSession({ ...session, nickname: e.target.value })
            }
          />
        </label>
        <div className="actions">
          <button className="primary" disabled={!ready} onClick={onCreate}>
            Я ОРГАНИЗАТОР
          </button>
          <button disabled={!ready} onClick={onJoin}>
            Я УЧАСТНИК
          </button>
        </div>
      </div>
      <div className="role-hints">
        <span>
          <b>ОРГАНИЗАТОР</b> создаёт лобби и показывает арену
        </span>
        <span>
          <b>УЧАСТНИК</b> входит по коду и решает задания
        </span>
      </div>
    </section>
  );
}

function Create({
  session,
  create,
  back,
}: {
  session: Session;
  create: (d: unknown) => void;
  back: () => void;
}) {
  const [form, setForm] = useState({
    serverName: "Турнир по информатике",
    maxPlayers: 16,
    gradeMode: "mixed",
    gameMode: "qualifier",
    settings: {
      roundDurationSeconds: 600,
      towerHp: 260,
      teamPlayerLimit: 4,
      zoneStepsToCenter: 8,
      zonePushbackSteps: 2,
      zoneHoldSeconds: 15,
    },
  });
  return (
    <Screen title="ОРГАНИЗАТОР // НОВОЕ ЛОББИ" back={back}>
      <div className="panel form-grid">
        <label>
          НАЗВАНИЕ
          <input
            value={form.serverName}
            onChange={(e) => setForm({ ...form, serverName: e.target.value })}
          />
        </label>
        <label>
          ТУР
          <select
            value={form.gameMode}
            onChange={(e) => setForm({ ...form, gameMode: e.target.value })}
          >
            <option value="qualifier">
              Отборочный: захват центральной зоны
            </option>
            <option value="final">Финал: команда против команды</option>
          </select>
        </label>
        {form.gameMode === "qualifier" ? (
          <label>
            УЧАСТНИКОВ В КАЖДОЙ ИЗ 8 КОМАНД
            <select
              value={form.settings.teamPlayerLimit}
              onChange={(e) =>
                setForm({
                  ...form,
                  settings: {
                    ...form.settings,
                    teamPlayerLimit: +e.target.value,
                  },
                })
              }
            >
              {[1, 2, 3, 4, 5, 6, 7, 8].map((n) => (
                <option key={n} value={n}>
                  {n} (всего мест: {n * 8})
                </option>
              ))}
            </select>
          </label>
        ) : (
          <label>
            МАКС. УЧАСТНИКОВ ФИНАЛА
            <select
              value={form.maxPlayers}
              onChange={(e) => setForm({ ...form, maxPlayers: +e.target.value })}
            >
              {[4, 6, 8, 10, 12, 16, 20, 24].map((n) => (
                <option key={n}>{n}</option>
              ))}
            </select>
          </label>
        )}
        <label>
          КЛАССЫ
          <select
            value={form.gradeMode}
            onChange={(e) => setForm({ ...form, gradeMode: e.target.value })}
          >
            <option value="9">9 класс</option>
            <option value="10">10 класс</option>
            <option value="11">11 класс</option>
            <option value="mixed">9–11 классы</option>
          </select>
        </label>
        <label>
          ДЛИТЕЛЬНОСТЬ, СЕК
          <input
            type="number"
            min="60"
            max="3600"
            value={form.settings.roundDurationSeconds}
            onChange={(e) =>
              setForm({
                ...form,
                settings: {
                  ...form.settings,
                  roundDurationSeconds: +e.target.value,
                },
              })
            }
          />
        </label>
        {form.gameMode === "final" && (
          <label>
            HP БАШЕН
            <input
              type="number"
              min="50"
              max="3000"
              value={form.settings.towerHp}
              onChange={(e) =>
                setForm({
                  ...form,
                  settings: { ...form.settings, towerHp: +e.target.value },
                })
              }
            />
          </label>
        )}
        {form.gameMode === "qualifier" && (
          <>
            <label>
              ШАГОВ ДО ЦЕНТРА
              <input
                type="number"
                min="4"
                max="20"
                value={form.settings.zoneStepsToCenter}
                onChange={(e) =>
                  setForm({
                    ...form,
                    settings: {
                      ...form.settings,
                      zoneStepsToCenter: +e.target.value,
                    },
                  })
                }
              />
            </label>
            <label>
              ОТБРОС ПРИ ВЫТЕСНЕНИИ
              <input
                type="number"
                min="1"
                max={Math.max(1, form.settings.zoneStepsToCenter - 1)}
                value={form.settings.zonePushbackSteps}
                onChange={(e) =>
                  setForm({
                    ...form,
                    settings: {
                      ...form.settings,
                      zonePushbackSteps: +e.target.value,
                    },
                  })
                }
              />
            </label>
            <label>
              УДЕРЖАНИЕ ЗОНЫ, СЕК
              <input
                type="number"
                min="5"
                max="120"
                value={form.settings.zoneHoldSeconds}
                onChange={(e) =>
                  setForm({
                    ...form,
                    settings: {
                      ...form.settings,
                      zoneHoldSeconds: +e.target.value,
                    },
                  })
                }
              />
            </label>
            <div className="qualifier-rule-note">
              <b>Важно:</b> в отборе всегда участвуют ровно 8 команд. Каждый
              участник выбирает команду после входа. Для запуска во всех восьми
              командах должен быть хотя бы один человек; в финал проходят 4 команды.
            </div>
          </>
        )}
        <button
          className="primary wide"
          onClick={() =>
            create({
              ...form,
              maxPlayers:
                form.gameMode === "qualifier"
                  ? form.settings.teamPlayerLimit * 8
                  : form.maxPlayers,
              nickname: session.nickname,
              grade: session.grade,
              settings: {
                ...form.settings,
                teamPlayerLimit:
                  form.gameMode === "qualifier"
                    ? form.settings.teamPlayerLimit
                    : Math.ceil(form.maxPlayers / 2),
              },
            })
          }
        >
          СОЗДАТЬ ЛОББИ
        </button>
      </div>
    </Screen>
  );
}

function Join({
  join,
  back,
}: {
  join: (id: string) => void;
  back: () => void;
}) {
  const [id, setId] = useState(roomIdFromPath());
  return (
    <Screen title="УЧАСТНИК // ВХОД ПО КОДУ" back={back}>
      <div className="panel join-card">
        <span>Введите код с экрана организатора</span>
        <input
          className="server-code-input"
          value={id}
          placeholder="CYB-XXXXX"
          onChange={(e) => setId(e.target.value.toUpperCase())}
        />
        <button
          className="primary"
          disabled={id.trim().length < 5}
          onClick={() => join(id)}
        >
          ПОДКЛЮЧИТЬСЯ
        </button>
      </div>
    </Screen>
  );
}

function Lobby({
  room,
  playerId,
  choose,
  chooseQualifier,
  start,
  finish,
}: {
  room: Room;
  playerId: string;
  choose: (t: TeamName) => void;
  chooseQualifier: (teamId: string) => void;
  start: () => void;
  finish: () => void;
}) {
  const me = room.players[playerId];
  const participants = Object.values(room.players).filter(
    (p) => p.role === "participant" && !p.isBot,
  );
  const qualifierTeams = Object.values(room.qualifierTeams || {}).sort(
    (a, b) => a.lane - b.lane,
  );
  const membersFor = (teamId: string) =>
    participants.filter((p) => p.qualifierTeamId === teamId);
  const emptyTeams = qualifierTeams.filter((team) => membersFor(team.id).length === 0);

  if (me?.role === "organizer")
    return (
      <Screen
        title={`ПАНЕЛЬ ОРГАНИЗАТОРА // ${room.serverName}`}
        badge={room.uniqueServerId}
      >
        <div className="organizer-lobby">
          <div className="panel lobby-code">
            <small>КОД ДЛЯ ПОДКЛЮЧЕНИЯ</small>
            <strong>{room.uniqueServerId}</strong>
            <p>
              Каждый игрок входит со своего устройства, указывает своё имя и
              выбирает одну из восьми команд.
            </p>
          </div>
          <div className="panel">
            <h2>{room.gameMode === "qualifier" ? "ЗАХВАТ ЗОНЫ // 8 КОМАНД" : "ФИНАЛ"}</h2>
            <div className="setting-list">
              <span>
                ПОДКЛЮЧЕНО <b>{participants.length}/{room.maxPlayers}</b>
              </span>
              <span>
                КЛАССЫ <b>{room.gradeMode}</b>
              </span>
              <span>
                ВРЕМЯ <b>{Math.floor(room.settings.roundDurationSeconds / 60)} мин</b>
              </span>
              {room.gameMode === "qualifier" && (
                <>
                  <span>
                    КОМАНД ЗАПОЛНЕНО <b>{8 - emptyTeams.length}/8</b>
                  </span>
                  <span>
                    МЕСТ В КОМАНДЕ <b>{room.settings.teamPlayerLimit}</b>
                  </span>
                  <span>
                    В ФИНАЛ ПРОЙДЁТ <b>4 команды</b>
                  </span>
                </>
              )}
            </div>
            {room.gameMode === "qualifier" && emptyTeams.length > 0 && (
              <p className="lobby-warning">
                Тур можно запустить сейчас. Пустые команды будут пропущены:
                {" "}
                {emptyTeams.map((team) => team.name).join(", ") || "нет"}.
              </p>
            )}
            <button
              className="primary pulse"
              disabled={false}
              onClick={start}
            >
              ЗАПУСТИТЬ ТУР
            </button>
            <button className="danger-action" onClick={finish}>
              ЗАВЕРШИТЬ ТУР
            </button>
          </div>
          <div className="panel participant-board team-roster-board">
            <h3>{room.gameMode === "qualifier" ? "СОСТАВЫ 8 КОМАНД" : "ПОДКЛЮЧИВШИЕСЯ УЧАСТНИКИ"}</h3>
            {room.gameMode === "qualifier" ? (
              <div className="organizer-team-grid">
                {qualifierTeams.map((team) => {
                  const members = membersFor(team.id);
                  return (
                    <div
                      key={team.id}
                      className="organizer-team-card"
                      style={{ "--team-hue": `${team.hue}` } as CSSProperties}
                    >
                      <div>
                        <b>{team.name}</b>
                        <span>{members.length}/{room.settings.teamPlayerLimit}</span>
                      </div>
                      {members.length ? (
                        <ul>
                          {members.map((player) => (
                            <li key={player.id}>{player.nickname} · {player.grade} кл.</li>
                          ))}
                        </ul>
                      ) : (
                        <small>Нет участников</small>
                      )}
                    </div>
                  );
                })}
              </div>
            ) : participants.length ? (
              participants.map((p) => (
                <ParticipantRow key={p.id} player={p} mode={room.gameMode} room={room} />
              ))
            ) : (
              <p className="waiting">Ожидание участников...</p>
            )}
          </div>
        </div>
      </Screen>
    );

  const selectedQualifierTeam = me?.qualifierTeamId
    ? room.qualifierTeams?.[me.qualifierTeamId]
    : undefined;

  return (
    <Screen
      title={`ЛОББИ УЧАСТНИКА // ${room.serverName}`}
      badge={room.uniqueServerId}
    >
      <div className="participant-lobby panel">
        <div className="ready-icon">✓</div>
        <h2>ВЫ ПОДКЛЮЧЕНЫ</h2>
        <p>Выберите команду. Организатор запустит тур, когда будут представлены все 8 команд.</p>
        {room.gameMode === "qualifier" ? (
          <div className="qualifier-team-select">
            <h3>ВЫБЕРИТЕ ОДНУ ИЗ 8 КОМАНД</h3>
            <div className="qualifier-team-grid">
              {qualifierTeams.map((team) => {
                const members = membersFor(team.id);
                const selected = me?.qualifierTeamId === team.id;
                const full = members.length >= room.settings.teamPlayerLimit && !selected;
                return (
                  <button
                    key={team.id}
                    className={`qualifier-team-choice ${selected ? "selected" : ""}`}
                    style={{ "--team-hue": `${team.hue}` } as CSSProperties}
                    disabled={full}
                    onClick={() => chooseQualifier(team.id)}
                  >
                    <b>{team.name}</b>
                    <span>{members.length}/{room.settings.teamPlayerLimit} участников</span>
                    <small>
                      {members.length
                        ? members.map((member) => member.nickname).join(", ")
                        : "Свободная команда"}
                    </small>
                  </button>
                );
              })}
            </div>
            <div className="selected-team-note">
              {selectedQualifierTeam
                ? `Вы в команде «${selectedQualifierTeam.name}». Все ответы участников двигают общий маркер команды.`
                : "Сначала выберите команду."}
            </div>
          </div>
        ) : (
          <div className="final-team-select">
            <h3>ВЫБЕРИТЕ СТОРОНУ</h3>
            <div>
              <button
                className={`nex-choice ${me?.team === "NexGen" ? "selected" : ""}`}
                onClick={() => choose("NexGen")}
              >
                NEXGEN
              </button>
              <button
                className={`omni-choice ${me?.team === "OmniSoft" ? "selected" : ""}`}
                onClick={() => choose("OmniSoft")}
              >
                OMNISOFT
              </button>
            </div>
          </div>
        )}
        <div className="mini-roster">
          Подключено: <b>{participants.length}/{room.maxPlayers}</b>
        </div>
      </div>
    </Screen>
  );
}

function ParticipantRow({
  player,
  mode,
  room,
}: {
  player: Player;
  mode: Room["gameMode"];
  room: Room;
}) {
  const qualifierTeam = player.qualifierTeamId
    ? room.qualifierTeams?.[player.qualifierTeamId]
    : undefined;
  return (
    <div className="participant-row">
      <span>{player.nickname.slice(0, 1).toUpperCase()}</span>
      <b>{player.nickname}</b>
      <small>{player.grade} класс</small>
      <em>{mode === "qualifier" ? qualifierTeam?.name || "не выбрана" : player.team || "ожидает"}</em>
    </div>
  );
}

function TournamentGame({
  room,
  playerId,
  questions,
  tasks,
  answer,
  submitTask,
  finish,
}: {
  room: Room;
  playerId: string;
  questions: Question[];
  tasks: TerminalTask[];
  answer: (q: string, a: number) => void;
  submitTask: (task: string, value: string) => void;
  finish: () => void;
}) {
  const me = room.players[playerId];
  return me?.role === "organizer" ? (
    <OrganizerArena room={room} finish={finish} />
  ) : (
    <ParticipantConsole
      room={room}
      player={me}
      questions={questions}
      tasks={tasks}
      answer={answer}
      submitTask={submitTask}
    />
  );
}

function laneY(lane: number, laneCount: number) {
  return laneCount <= 1 ? 50 : 16 + (lane * 68) / (laneCount - 1);
}
function routePoint(lane: number, position: number, laneCount: number) {
  const y = lane < 0 ? 50 : laneY(lane, laneCount);
  if (position <= 18)
    return { x: position, y: 50 + ((y - 50) * Math.max(0, position - 4)) / 14 };
  if (position >= 82)
    return { x: position, y: y + (50 - y) * Math.min(1, (position - 82) / 14) };
  return { x: position, y };
}

function OrganizerArena({ room, finish }: { room: Room; finish: () => void }) {
  return room.gameMode === "qualifier" ? (
    <QualifierArena room={room} finish={finish} />
  ) : (
    <FinalArena room={room} finish={finish} />
  );
}

function radialPoint(lane: number, progress: number, count: number) {
  const angle = -Math.PI / 2 + (Math.PI * 2 * lane) / Math.max(1, count);
  const radius = 43 * (1 - Math.min(100, Math.max(0, progress)) / 100);
  return { x: 50 + Math.cos(angle) * radius, y: 50 + Math.sin(angle) * radius };
}

function qualifierStatusLabel(status: string) {
  if (status === "holding") return "В ЗОНЕ";
  if (status === "qualified") return "ПРОШЛА В ФИНАЛ";
  if (status === "eliminated") return "ВЫБЫЛА";
  return "ДВИЖЕТСЯ";
}

function QualifierArena({ room, finish }: { room: Room; finish: () => void }) {
  const participants = Object.values(room.players).filter(
    (p) => p.role === "participant" && !p.isBot,
  );
  const teams = Object.values(room.qualifierTeams || {}).sort(
    (a, b) => a.lane - b.lane,
  );
  const membersFor = (teamId: string) =>
    participants.filter((player) => player.qualifierTeamId === teamId);
  const holder = room.zoneHolderTeamId
    ? room.qualifierTeams?.[room.zoneHolderTeamId]
    : undefined;
  const remaining = Math.max(
    0,
    (room.endsAt || 0) - Math.floor(Date.now() / 1000),
  );
  const qualified = new Set(room.qualifiedTeamIds || []);
  const holdPct = holder
    ? Math.min(
        100,
        (holder.captureProgress / Math.max(1, room.settings.zoneHoldSeconds)) *
          100,
      )
    : 0;
  return (
    <section className="game-screen organizer-arena">
      <div className="projector-banner">
        <b>ОТБОРОЧНЫЙ ТУР // 8 КОМАНД // ЗАХВАТ ЗОНЫ</b>
        <span>{room.uniqueServerId}</span>
        <button className="danger-action compact" onClick={finish}>
          ЗАВЕРШИТЬ
        </button>
      </div>
      <div className="zone-top">
        <div className="zone-summary">
          <small>В ФИНАЛЕ</small>
          <b>{qualified.size} / 4</b>
        </div>
        <div className="timer">
          {formatTime(remaining)}
          <small>ОСТАЛОСЬ</small>
        </div>
        <div className={`zone-holder-card ${holder ? "occupied" : ""}`}>
          <small>ЦЕНТРАЛЬНАЯ ЗОНА</small>
          <b>{holder?.name || "СВОБОДНА"}</b>
          <i>
            <em style={{ width: `${holdPct}%` }} />
          </i>
          <span>
            {holder
              ? `${holder.captureProgress}/${room.settings.zoneHoldSeconds} сек`
              : "Команды движутся к центру"}
          </span>
        </div>
      </div>
      <div className="story">{room.storyMessage}</div>
      <div className="arena zone-arena">
        <svg
          className="zone-spokes"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          {teams.map((team) => {
            const point = radialPoint(team.lane, 0, teams.length);
            return (
              <line key={team.id} x1={point.x} y1={point.y} x2="50" y2="50" />
            );
          })}
        </svg>
        <div className="zone-ring ring-outer" />
        <div className="zone-ring ring-middle" />
        <div className={`capture-zone ${holder ? "occupied" : ""}`}>
          <span>CAPTURE</span>
          <b>ZONE</b>
          <i>
            <em style={{ width: `${holdPct}%` }} />
          </i>
        </div>
        {teams
          .filter((team) => !qualified.has(team.id))
          .map((team) => {
            const progress =
              (team.zoneSteps / Math.max(1, room.settings.zoneStepsToCenter)) *
              100;
            const point = radialPoint(team.lane, progress, teams.length);
            const members = membersFor(team.id);
            return (
              <div
                key={team.id}
                className={`zone-team ${team.status || "active"}`}
                style={
                  {
                    left: `${point.x}%`,
                    top: `${point.y}%`,
                    "--team-hue": `${team.hue}`,
                  } as CSSProperties
                }
              >
                <span>{team.name.slice(0, 2).toUpperCase()}</span>
                <b>{team.name}</b>
                <small>
                  {team.zoneSteps}/{room.settings.zoneStepsToCenter} · {members.length} чел.
                </small>
              </div>
            );
          })}
        <div className="qualified-rail">
          <small>ПРОШЛИ В ФИНАЛ</small>
          {(room.qualifiedTeamIds || []).map((id, index) => (
            <div key={id}>
              <b>#{index + 1}</b>
              <span>{room.qualifierTeams?.[id]?.name}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="organizer-bottom">
        <div className="panel event-feed">
          <small>ПОСЛЕДНЕЕ СОБЫТИЕ</small>
          <strong>{room.lastEvent || "Синхронизация..."}</strong>
        </div>
        <div className="panel live-roster">
          <h3>ПОЛОЖЕНИЕ 8 КОМАНД</h3>
          {[...teams]
            .sort(
              (a, b) =>
                b.zoneSteps - a.zoneSteps ||
                b.captureProgress - a.captureProgress ||
                b.score - a.score,
            )
            .map((team) => (
              <div key={team.id}>
                <b>{team.name}</b>
                <span>
                  {qualifierStatusLabel(team.status)} // {membersFor(team.id).length} участников
                </span>
                <strong>
                  {team.zoneSteps}/{room.settings.zoneStepsToCenter} · {team.score}
                </strong>
              </div>
            ))}
        </div>
      </div>
    </section>
  );
}

function FinalArena({ room, finish }: { room: Room; finish: () => void }) {
  const units = Object.values(room.units);
  const laneCount = Math.max(1, ...units.map((u) => u.lane + 1));
  const remaining = Math.max(
    0,
    (room.endsAt || 0) - Math.floor(Date.now() / 1000),
  );
  const participants = Object.values(room.players)
    .filter((p) => p.role === "participant" && !p.isBot)
    .sort((a, b) => b.score - a.score);
  return (
    <section className="game-screen organizer-arena">
      <div className="projector-banner">
        <b>ФИНАЛ // КОМАНДА ПРОТИВ КОМАНДЫ</b>
        <span>{room.uniqueServerId}</span>
        <button className="danger-action compact" onClick={finish}>
          ЗАВЕРШИТЬ
        </button>
      </div>
      <div className="game-top">
        <Tower team={room.teams.NexGen} label="NEXGEN TOWER" />
        <div className="timer">
          {formatTime(remaining)}
          <small>ОСТАЛОСЬ</small>
        </div>
        <Tower team={room.teams.OmniSoft} label="OMNISOFT TOWER" />
      </div>
      <div className="story">{room.storyMessage}</div>
      <div className="arena topdown">
        <svg
          className="route-map"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
        >
          {Array.from({ length: laneCount }, (_, lane) => {
            const y = laneY(lane, laneCount);
            return (
              <g key={lane}>
                <polyline points={`4,50 18,${y} 82,${y} 96,50`} />
                {[28, 50, 72].map((x) => (
                  <circle key={`${lane}-${x}`} cx={x} cy={y} r="0.7" />
                ))}
              </g>
            );
          })}
        </svg>
        <div className="tower-icon left">
          <img src="/assets/openmoji/tower.svg" alt="башня NexGen" />
          <b>NEX</b>
        </div>
        <div className="tower-icon right">
          <img src="/assets/openmoji/tower.svg" alt="башня OmniSoft" />
          <b>OMNI</b>
        </div>
        {units.map((u) => (
          <ArenaUnit key={u.ownerPlayerId} unit={u} laneCount={laneCount} />
        ))}
        {(room.projectiles || []).map((p) => {
          const from = routePoint(p.fromLane, p.from, laneCount);
          const to = routePoint(p.toLane, p.to, laneCount);
          const angle =
            (Math.atan2(to.y - from.y, to.x - from.x) * 180) / Math.PI;
          return (
            <div
              key={p.id}
              className={`shot ${p.team === "NexGen" ? "nex" : "omni"}`}
              style={
                {
                  left: `${from.x}%`,
                  top: `${from.y}%`,
                  "--shot-to-x": `${to.x}%`,
                  "--shot-to-y": `${to.y}%`,
                  "--shot-angle": `${angle + 90}deg`,
                } as CSSProperties
              }
            >
              <img src="/assets/openmoji/rocket.svg" alt="снаряд" />
            </div>
          );
        })}
      </div>
      <div className="organizer-bottom">
        <div className="panel event-feed">
          <small>ПОСЛЕДНЕЕ СОБЫТИЕ</small>
          <strong>{room.lastEvent || "Синхронизация..."}</strong>
        </div>
        <div className="panel live-roster">
          <h3>LIVE SCORE</h3>
          {participants.map((p) => (
            <div key={p.id}>
              <b>{p.nickname}</b>
              <span>{p.latestBuff || "—"}</span>
              <strong>{p.score}</strong>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function ArenaUnit({ unit, laneCount }: { unit: Unit; laneCount: number }) {
  const point = routePoint(unit.lane, unit.position, laneCount);
  return (
    <div
      className={`unit ${unit.team === "NexGen" ? "nex" : "omni"} ${unit.hp <= 0 ? "down" : ""} ${unit.isBoss ? "boss-unit" : ""}`}
      style={{ left: `${point.x}%`, top: `${point.y}%` }}
    >
      <b>{unit.nickname}</b>
      <img src="/assets/openmoji/robot.svg" alt="боевой юнит" />
      <i>
        <em
          style={{ width: `${Math.max(0, (unit.hp / unit.maxHp) * 100)}%` }}
        />
      </i>
      <small>
        LVL {unit.level} // {Math.max(0, unit.hp)} HP
      </small>
    </div>
  );
}

function ParticipantConsole({
  room,
  player,
  questions,
  tasks,
  answer,
  submitTask,
}: {
  room: Room;
  player?: Player;
  questions: Question[];
  tasks: TerminalTask[];
  answer: (q: string, a: number) => void;
  submitTask: (task: string, value: string) => void;
}) {
  const [tab, setTab] = useState<"theory" | "code">("theory");
  const [taskAnswer, setTaskAnswer] = useState<Record<string, string>>({});
  const [selectedTask, setSelectedTask] = useState("");
  useEffect(() => {
    if (!selectedTask && tasks[0]) setSelectedTask(tasks[0].id);
  }, [tasks, selectedTask]);
  const question = questions.find((q) => q.id === player?.questionId);
  const task = tasks.find((t) => t.id === selectedTask) || tasks[0];
  const battleUnit = player ? room.units[player.id] : undefined;
  const qualifierTeam = player?.qualifierTeamId
    ? room.qualifierTeams?.[player.qualifierTeamId]
    : undefined;
  const remaining = Math.max(
    0,
    (room.endsAt || 0) - Math.floor(Date.now() / 1000),
  );
  const locked = (player?.lockedUntil || 0) > Date.now() / 1000;
  const qualifierDone =
    room.gameMode === "qualifier" &&
    (qualifierTeam?.status === "qualified" ||
      qualifierTeam?.status === "eliminated");
  const holder = room.zoneHolderTeamId === qualifierTeam?.id;
  const stepsRemaining = Math.max(
    0,
    room.settings.zoneStepsToCenter - (qualifierTeam?.zoneSteps || 0),
  );
  const statusText =
    qualifierTeam?.status === "qualified"
      ? "ВАША КОМАНДА ПРОШЛА В ФИНАЛ"
      : qualifierTeam?.status === "eliminated"
        ? "ОТБОРОЧНЫЙ ТУР ЗАВЕРШЁН"
        : holder
          ? "ВАША КОМАНДА УДЕРЖИВАЕТ ЗОНУ"
          : `ДО ЗОНЫ: ${stepsRemaining} ШАГОВ`;
  return (
    <section className="participant-console">
      <div className="participant-top">
        <div>
          <small>
            {room.gameMode === "qualifier"
              ? `ОТБОР // ${qualifierTeam?.name || "КОМАНДА НЕ ВЫБРАНА"}`
              : `ФИНАЛ // ${player?.team}`}
          </small>
          <h1>{player?.nickname}</h1>
        </div>
        <div className="personal-timer">{formatTime(remaining)}</div>
      </div>
      <div className={`buff-strip ${holder ? "zone-owned" : ""}`}>
        <span>
          {room.gameMode === "qualifier" ? "СТАТУС КОМАНДЫ" : "ПОСЛЕДНИЙ БАФФ"}
        </span>
        <b>
          {room.gameMode === "qualifier"
            ? `${statusText} // ${player?.latestBuff || "Решайте задания вместе"}`
            : player?.latestBuff || "Решите задание, чтобы усилить бойца"}
        </b>
      </div>
      {room.gameMode === "qualifier" ? (
        <div className="participant-stats">
          <Stat
            name="ШАГИ"
            value={`${qualifierTeam?.zoneSteps || 0}/${room.settings.zoneStepsToCenter}`}
          />
          <Stat
            name="ЗАХВАТ"
            value={`${qualifierTeam?.captureProgress || 0}/${room.settings.zoneHoldSeconds}`}
          />
          <Stat name="ВЕРНО" value={player?.correctAnswers} />
          <Stat name="КОД" value={player?.solvedTasks?.length || 0} />
          <Stat name="ЛИЧНЫЙ SCORE" value={player?.score} />
          <Stat name="КОМАНДНЫЙ SCORE" value={qualifierTeam?.score || 0} />
        </div>
      ) : (
        <div className="participant-stats">
          <Stat name="LEVEL" value={player?.level} />
          <Stat name="ATTACK" value={player?.attack} />
          <Stat name="DEFENCE" value={player?.defense} />
          <Stat
            name="HP"
            value={`${Math.max(0, battleUnit?.hp ?? player?.hp ?? 0)}/${battleUnit?.maxHp ?? player?.maxHp ?? 0}`}
          />
          <Stat name="SCORE" value={player?.score} />
        </div>
      )}
      <div className="panel challenge-panel">
        {qualifierDone ? (
          <div className={`qualifier-finish ${qualifierTeam?.status}`}>
            <b>
              {qualifierTeam?.status === "qualified"
                ? "ФИНАЛ ДОСТУПЕН"
                : "СПАСИБО ЗА ИГРУ"}
            </b>
            <span>
              {qualifierTeam?.status === "qualified"
                ? `Команда «${qualifierTeam?.name}» вошла в число четырёх сильнейших. Ожидайте финал.`
                : "В следующий тур прошли четыре из восьми команд. Результаты видны на экране организатора."}
            </span>
          </div>
        ) : (
          <>
            <div className="tabs">
              <button
                className={tab === "theory" ? "active" : ""}
                onClick={() => setTab("theory")}
              >
                ТЕОРИЯ + ЛОГИКА
              </button>
              <button
                className={tab === "code" ? "active" : ""}
                onClick={() => setTab("code")}
              >
                НАПИСАНИЕ КОДА
              </button>
            </div>
            {tab === "theory" ? (
              <div className="theory-workspace">
                {locked ? (
                  <div className="locked">
                    <b>МОДУЛЬ ЗАБЛОКИРОВАН</b>
                    <span>
                      После трёх ошибок сделайте паузу и проверьте рассуждение.
                    </span>
                  </div>
                ) : question ? (
                  <>
                    <div className="question-head">
                      <span>{question.topic}</span>
                      <b>СЛОЖНОСТЬ {question.difficulty}/3</b>
                    </div>
                    <h2>{question.text}</h2>
                    <div className="options">
                      {question.options.map((o, i) => (
                        <button
                          key={`${question.id}-${i}`}
                          onClick={() => answer(question.id, i)}
                        >
                          <kbd>{String.fromCharCode(65 + i)}</kbd>
                          {o}
                        </button>
                      ))}
                    </div>
                  </>
                ) : (
                  <p>Получаем следующее задание...</p>
                )}
              </div>
            ) : (
              <div className="code-workspace">
                <aside>
                  {tasks.map((t) => (
                    <button
                      key={t.id}
                      className={`${selectedTask === t.id ? "active" : ""} ${player?.solvedTasks?.includes(t.id) ? "solved" : ""}`}
                      onClick={() => setSelectedTask(t.id)}
                    >
                      <b>{t.title}</b>
                      <small>
                        {t.language} // +{t.reward}
                      </small>
                    </button>
                  ))}
                </aside>
                {task && (
                  <div className="code-task">
                    <div className="question-head">
                      <span>{task.language.toUpperCase()}</span>
                      <b>СЛОЖНОСТЬ {task.difficulty}/3</b>
                    </div>
                    <h2>{task.title}</h2>
                    <p>{task.description}</p>
                    {task.starterCode && <pre>{task.starterCode}</pre>}
                    <textarea
                      spellCheck={false}
                      disabled={player?.solvedTasks?.includes(task.id)}
                      value={taskAnswer[task.id] || ""}
                      placeholder="Введите недостающий код или запрос..."
                      onChange={(e) =>
                        setTaskAnswer({
                          ...taskAnswer,
                          [task.id]: e.target.value,
                        })
                      }
                    />
                    <button
                      className="primary"
                      disabled={
                        player?.solvedTasks?.includes(task.id) ||
                        !(taskAnswer[task.id] || "").trim()
                      }
                      onClick={() =>
                        submitTask(task.id, taskAnswer[task.id] || "")
                      }
                    >
                      {player?.solvedTasks?.includes(task.id)
                        ? "ЗАДАЧА РЕШЕНА"
                        : "ОТПРАВИТЬ НА ПРОВЕРКУ"}
                    </button>
                    <small className="code-note">
                      Код не запускается на компьютере организатора: MVP сверяет
                      безопасные варианты ответа.
                    </small>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </section>
  );
}

function Tower({
  team,
  label,
}: {
  team: { name: TeamName; towerHp: number; maxTowerHp: number };
  label?: string;
}) {
  const pct = Math.max(0, (team.towerHp / team.maxTowerHp) * 100);
  return (
    <div className={`tower-hp ${team.name === "NexGen" ? "nex" : "omni"}`}>
      <div>
        <b>{label || `${team.name} TOWER`}</b>
        <span>
          {Math.max(0, team.towerHp)} / {team.maxTowerHp}
        </span>
      </div>
      <i>
        <em style={{ width: `${pct}%` }} />
      </i>
    </div>
  );
}

function Stat({
  name,
  value,
}: {
  name: string;
  value: string | number | undefined;
}) {
  return (
    <div className="stat">
      <span>{name}</span>
      <b>{value}</b>
    </div>
  );
}
function formatTime(seconds: number) {
  return `${Math.floor(seconds / 60)
    .toString()
    .padStart(2, "0")}:${(seconds % 60).toString().padStart(2, "0")}`;
}

function Results({ room, restart }: { room: Room; restart: () => void }) {
  const players = useMemo(
    () =>
      Object.values(room.players)
        .filter((p) => p.role === "participant" && !p.isBot)
        .sort((a, b) => b.score - a.score),
    [room],
  );
  const qualifierTeams = useMemo(
    () =>
      Object.values(room.qualifierTeams || {}).sort((a, b) => {
        const aq = a.status === "qualified" ? 1 : 0;
        const bq = b.status === "qualified" ? 1 : 0;
        return (
          bq - aq ||
          b.zoneSteps - a.zoneSteps ||
          b.captureProgress - a.captureProgress ||
          b.score - a.score
        );
      }),
    [room],
  );
  const membersFor = (teamId: string) =>
    players.filter((player) => player.qualifierTeamId === teamId);

  if (room.gameMode === "qualifier")
    return (
      <Screen title="ОТБОРОЧНЫЙ ТУР // РЕЗУЛЬТАТЫ">
        <div className="winner blue-text">
          <small>В ФИНАЛ ПРОШЛИ</small>4 КОМАНДЫ
        </div>
        <p className="result-story">{room.storyMessage}</p>
        <div className="panel leaderboard qualifier-results">
          {qualifierTeams.map((team, i) => (
            <div key={team.id} className={team.status || "eliminated"}>
              <b>#{i + 1}</b>
              <span>
                {team.name}
                <small>
                  {membersFor(team.id).map((player) => player.nickname).join(", ")}
                </small>
              </span>
              <span>{team.status === "qualified" ? "ФИНАЛ" : "ВЫБЫЛА"}</span>
              <span>
                {team.zoneSteps}/{room.settings.zoneStepsToCenter} шагов
              </span>
              <strong>{team.score}</strong>
            </div>
          ))}
        </div>
        <button className="primary center" onClick={restart}>
          НОВОЕ ЛОББИ
        </button>
      </Screen>
    );
  return (
    <Screen title="ФИНАЛ ЗАВЕРШЁН // РЕЗУЛЬТАТЫ">
      <div
        className={`winner ${room.winner === "NexGen" ? "blue-text" : "pink-text"}`}
      >
        <small>ПОБЕДИТЕЛЬ</small>
        {room.winner}
      </div>
      <p className="result-story">{room.storyMessage}</p>
      <div className="panel leaderboard">
        {players.map((p, i) => (
          <div key={p.id}>
            <b>#{i + 1}</b>
            <span>
              {p.nickname}
              <small>{p.team}</small>
            </span>
            <span>{p.correctAnswers} верно</span>
            <span>{p.solvedTasks.length} код</span>
            <strong>{p.score}</strong>
          </div>
        ))}
      </div>
      <button className="primary center" onClick={restart}>
        НОВОЕ ЛОББИ
      </button>
    </Screen>
  );
}

function Screen({
  title,
  badge,
  back,
  children,
}: {
  title: string;
  badge?: string;
  back?: () => void;
  children: ReactNode;
}) {
  return (
    <section className="screen">
      <div className="screen-title">
        {back && <button onClick={back}>← НАЗАД</button>}
        <h1>{title}</h1>
        {badge && <strong>{badge}</strong>}
      </div>
      {children}
    </section>
  );
}
