import { FormEvent, useEffect, useMemo, useState } from "react";
import { Activity, ArrowRight, Check, CircleHelp, ClipboardList, FolderKanban, LogOut, Plus, Search, ShieldCheck, Sparkles, X } from "lucide-react";

const API = import.meta.env.VITE_API_URL || "http://localhost:8080";
type User = { id: number; username: string; email: string; role: string };
type Project = { id: number; name: string; description: string; owner_id: number; created_at: string };
type Task = { id: number; title: string; description: string; status: string; priority: string; project_id: number; deadline?: string; assignee_id?: number };
type Comment = { id: number; text: string; author_id: number; created_at: string };

async function request<T>(path: string, token: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}), ...init.headers },
  });
  if (response.status === 204) return undefined as T;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error || body.detail || `Request failed (${response.status})`);
  return body as T;
}

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem("tm-token") || "");
  const [user, setUser] = useState<User | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [comments, setComments] = useState<Comment[]>([]);
  const [selectedTask, setSelectedTask] = useState<number | null>(null);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [authMode, setAuthMode] = useState<"login" | "register">("login");
  const [auth, setAuth] = useState({ username: "", email: "", password: "" });
  const [projectForm, setProjectForm] = useState({ name: "", description: "" });
  const [taskForm, setTaskForm] = useState({ title: "", description: "", project_id: "", priority: "medium", status: "new" });
  const [commentText, setCommentText] = useState("");
  const [showProjectForm, setShowProjectForm] = useState(false);
  const [showTaskForm, setShowTaskForm] = useState(false);

  async function loadWorkspace(activeToken = token) {
    const [me, projectList, taskList] = await Promise.all([
      request<User>("/users/me", activeToken),
      request<Project[]>("/projects/", activeToken),
      request<Task[]>("/tasks/", activeToken),
    ]);
    setUser(me); setProjects(projectList); setTasks(taskList);
    if (!taskForm.project_id && projectList[0]) setTaskForm(v => ({ ...v, project_id: String(projectList[0].id) }));
  }

  useEffect(() => {
    if (!token) return;
    loadWorkspace().catch(e => { setError(e.message); if (/token|unauthorized|user not found/i.test(e.message)) logout(); });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  useEffect(() => {
    if (!selectedTask || !token) { setComments([]); return; }
    request<Comment[]>(`/comments/task/${selectedTask}`, token).then(setComments).catch(e => setError(e.message));
  }, [selectedTask, token]);

  function logout() {
    localStorage.removeItem("tm-token"); setToken(""); setUser(null); setProjects([]); setTasks([]); setSelectedTask(null);
  }

  async function handleAuth(e: FormEvent) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      if (authMode === "register") {
        await request<User>("/auth/register", "", { method: "POST", body: JSON.stringify(auth) });
      }
      const data = await request<{ access_token: string }>("/auth/login", "", {
        method: "POST", body: JSON.stringify({ username: auth.username, password: auth.password }),
      });
      localStorage.setItem("tm-token", data.access_token); setToken(data.access_token);
    } catch (e) { setError(e instanceof Error ? e.message : "Не удалось войти"); }
    finally { setBusy(false); }
  }

  async function createProject(e: FormEvent) {
    e.preventDefault(); setError("");
    try {
      await request<Project>("/projects/", token, { method: "POST", body: JSON.stringify(projectForm) });
      setProjectForm({ name: "", description: "" }); setShowProjectForm(false); await loadWorkspace();
    } catch (e) { setError(e instanceof Error ? e.message : "Не удалось создать проект"); }
  }

  async function createTask(e: FormEvent) {
    e.preventDefault(); setError("");
    try {
      await request<Task>("/tasks/", token, { method: "POST", body: JSON.stringify({ ...taskForm, project_id: Number(taskForm.project_id) }) });
      setTaskForm(v => ({ ...v, title: "", description: "" })); setShowTaskForm(false); await loadWorkspace();
    } catch (e) { setError(e instanceof Error ? e.message : "Не удалось создать задачу"); }
  }

  async function addComment(e: FormEvent) {
    e.preventDefault(); if (!selectedTask || !commentText.trim()) return;
    try {
      await request<Comment>("/comments/", token, { method: "POST", body: JSON.stringify({ task_id: selectedTask, text: commentText }) });
      setCommentText(""); setComments(await request<Comment[]>(`/comments/task/${selectedTask}`, token));
    } catch (e) { setError(e instanceof Error ? e.message : "Не удалось добавить комментарий"); }
  }

  async function changeStatus(task: Task, status: string) {
    try {
      await request<Task>(`/tasks/${task.id}`, token, { method: "PUT", body: JSON.stringify({ title: task.title, description: task.description, status, priority: task.priority, deadline: task.deadline || null, project_id: task.project_id, assignee_id: task.assignee_id || null }) });
      await loadWorkspace();
    } catch (e) { setError(e instanceof Error ? e.message : "Не удалось обновить задачу"); }
  }

  const filteredTasks = useMemo(() => tasks.filter(task => {
    const matchesQuery = `${task.title} ${task.description}`.toLowerCase().includes(query.toLowerCase());
    return matchesQuery && (statusFilter === "all" || task.status === statusFilter);
  }), [tasks, query, statusFilter]);
  const selected = tasks.find(t => t.id === selectedTask);
  const completed = tasks.filter(t => t.status === "completed").length;
  const progress = tasks.length ? Math.round(completed / tasks.length * 100) : 0;

  if (!token || !user) return <main className="auth-page">
    <div className="auth-glow glow-one" /><div className="auth-glow glow-two" />
    <section className="auth-card">
      <div className="brand"><div className="brand-mark"><Sparkles size={21} /></div><span>task<span className="brand-accent">manager</span></span></div>
      <div className="eyebrow">WORKSPACE / REWORK</div>
      <h1>Планы становятся<br /><span>результатами.</span></h1>
      <p className="muted auth-copy">Организуй проекты, следи за задачами и держи фокус на важном.</p>
      <div className="auth-tabs"><button className={authMode === "login" ? "active" : ""} onClick={() => setAuthMode("login")}>Войти</button><button className={authMode === "register" ? "active" : ""} onClick={() => setAuthMode("register")}>Создать аккаунт</button></div>
      <form onSubmit={handleAuth} className="form-stack">
        <label>Имя пользователя<input autoComplete="username" value={auth.username} onChange={e => setAuth({ ...auth, username: e.target.value })} required minLength={3} placeholder="например, avernus" /></label>
        {authMode === "register" && <label>Email<input type="email" autoComplete="email" value={auth.email} onChange={e => setAuth({ ...auth, email: e.target.value })} required placeholder="you@example.com" /></label>}
        <label>Пароль<input type="password" autoComplete={authMode === "login" ? "current-password" : "new-password"} value={auth.password} onChange={e => setAuth({ ...auth, password: e.target.value })} required minLength={8} placeholder="Минимум 8 символов" /></label>
        {error && <div className="error">{error}</div>}
        <button className="primary-btn full" disabled={busy}>{busy ? "Подождите…" : authMode === "login" ? "Войти в workspace" : "Создать аккаунт"}<ArrowRight size={17} /></button>
      </form>
      <div className="auth-foot"><ShieldCheck size={15} /> Защищённая авторизация · Go + PostgreSQL</div>
    </section>
  </main>;

  return <div className="app-shell">
    <aside className="sidebar">
      <div className="brand sidebar-brand"><div className="brand-mark"><Sparkles size={19} /></div><span>task<span className="brand-accent">manager</span></span></div>
      <div className="workspace-label">WORKSPACE <span className="status-dot" /></div>
      <button className="nav-item selected"><FolderKanban size={17} /> Обзор <span className="nav-count">{projects.length}</span></button>
      <button className="nav-item" onClick={() => setShowTaskForm(true)}><ClipboardList size={17} /> Все задачи <span className="nav-count">{tasks.length}</span></button>
      <div className="sidebar-section-title">МОИ ПРОЕКТЫ <button onClick={() => setShowProjectForm(true)} aria-label="Добавить проект"><Plus size={15} /></button></div>
      <div className="project-nav">{projects.map((p, i) => <button key={p.id} className="project-nav-item" onClick={() => { setTaskForm(v => ({ ...v, project_id: String(p.id) })); setQuery(""); }}><span className={`project-dot dot-${i % 4}`} />{p.name}</button>)}</div>
      <div className="sidebar-bottom"><div className="help-link"><CircleHelp size={16} /> Документация API <span>↗</span></div><div className="profile"><div className="avatar">{user.username.slice(0, 1).toUpperCase()}</div><div className="profile-copy"><strong>{user.username}</strong><small>{user.role === "admin" ? "Администратор" : "Участник"}</small></div><button onClick={logout} title="Выйти"><LogOut size={16} /></button></div></div>
    </aside>
    <main className="main-area">
      <header className="topbar"><div className="breadcrumb">Workspace <span>/</span> <strong>Обзор</strong></div><div className="topbar-right"><span className="live-badge"><span /> API connected</span><div className="avatar small">{user.username.slice(0, 1).toUpperCase()}</div></div></header>
      <div className="page-content">
        <section className="welcome-row"><div><div className="eyebrow">ПОНЕДЕЛЬНИК — НОВЫЙ СТАРТ</div><h1>Добро пожаловать, {user.username}<span className="wave"> ✦</span></h1><p className="muted">Вот что происходит в твоём workspace прямо сейчас.</p></div><button className="primary-btn" onClick={() => setShowTaskForm(true)}><Plus size={17} /> Новая задача</button></section>
        {error && <div className="error banner-error">{error}<button onClick={() => setError("")}><X size={16} /></button></div>}
        <section className="stats-grid">
          <div className="stat-card"><div className="stat-top"><span>Всего задач</span><div className="stat-icon purple"><ClipboardList size={18} /></div></div><div className="stat-number">{tasks.length}<span className="stat-context">в workspace</span></div><div className="stat-bottom"><span className="mini-bar"><i style={{ width: `${progress}%` }} /></span><span>{progress}% выполнено</span></div></div>
          <div className="stat-card"><div className="stat-top"><span>Активные проекты</span><div className="stat-icon blue"><FolderKanban size={18} /></div></div><div className="stat-number">{projects.length}<span className="stat-context">проектов</span></div><div className="stat-note">Всё организовано в одном месте</div></div>
          <div className="stat-card"><div className="stat-top"><span>Завершено</span><div className="stat-icon green"><Check size={18} /></div></div><div className="stat-number">{completed}<span className="stat-context">задач</span></div><div className="stat-note positive"><Activity size={14} /> Отличный прогресс</div></div>
        </section>
        <section className="content-grid">
          <div className="panel tasks-panel">
            <div className="panel-heading"><div><h2>Твои задачи</h2><p className="muted">Управляй приоритетами и двигайся вперёд.</p></div><button className="outline-btn" onClick={() => setShowTaskForm(true)}><Plus size={15} /> Добавить</button></div>
            <div className="toolbar"><div className="searchbox"><Search size={16} /><input value={query} onChange={e => setQuery(e.target.value)} placeholder="Поиск задач..." /></div><select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}><option value="all">Все статусы</option><option value="new">Новые</option><option value="in_progress">В работе</option><option value="completed">Готово</option></select></div>
            <div className="task-list">{filteredTasks.map(task => {
              const project = projects.find(p => p.id === task.project_id);
              return <article key={task.id} className={`task-row ${selectedTask === task.id ? "task-selected" : ""}`} onClick={() => setSelectedTask(task.id)}>
                <button className={`task-check ${task.status === "completed" ? "checked" : ""}`} title="Переключить статус" onClick={e => { e.stopPropagation(); changeStatus(task, task.status === "completed" ? "new" : "completed"); }}>{task.status === "completed" && <Check size={13} />}</button>
                <div className="task-main"><strong className={task.status === "completed" ? "task-done" : ""}>{task.title}</strong><div className="task-meta"><span className="project-pill"><span className="project-dot dot-1" />{project?.name || `Project ${task.project_id}`}</span>{task.deadline && <span>{new Date(task.deadline).toLocaleDateString("ru-RU")}</span>}</div></div>
                <span className={`priority priority-${task.priority}`}>{task.priority === "high" ? "Высокий" : task.priority === "low" ? "Низкий" : "Средний"}</span>
                <select className={`status-select status-${task.status}`} value={task.status} onClick={e => e.stopPropagation()} onChange={e => changeStatus(task, e.target.value)}><option value="new">Новая</option><option value="in_progress">В работе</option><option value="completed">Готово</option></select>
              </article>;
            })}
            {!filteredTasks.length && <div className="empty-state"><div className="empty-icon"><ClipboardList size={22} /></div><strong>{tasks.length ? "Ничего не найдено" : "Пока нет задач"}</strong><p>{tasks.length ? "Попробуй изменить фильтры." : "Создай первую задачу, чтобы начать работу."}</p>{!tasks.length && <button className="primary-btn" onClick={() => setShowTaskForm(true)}><Plus size={16} /> Создать задачу</button>}</div>}
            </div>
            <div className="panel-footer"><span>Показано {filteredTasks.length} из {tasks.length} задач</span><span className="secure-note"><ShieldCheck size={14} /> Доступ контролируется сервером</span></div>
          </div>
          <div className="right-column">
            <div className="panel projects-panel"><div className="panel-heading"><div><h2>Проекты</h2><p className="muted">Твои рабочие пространства.</p></div><button className="icon-btn" onClick={() => setShowProjectForm(true)} aria-label="Добавить проект"><Plus size={18} /></button></div>
              {projects.slice(0, 4).map((project, i) => { const count = tasks.filter(t => t.project_id === project.id).length; return <div className="project-card" key={project.id}><div className={`project-symbol symbol-${i % 4}`}><FolderKanban size={18} /></div><div className="project-card-copy"><strong>{project.name}</strong><span>{count} задач · {project.description || "Без описания"}</span><div className="project-progress"><i style={{ width: `${count ? tasks.filter(t => t.project_id === project.id && t.status === "completed").length / count * 100 : 0}%` }} /></div></div><span className="project-arrow">↗</span></div> })}
              {!projects.length && <div className="small-empty">Создай проект, чтобы собрать задачи в одном месте.</div>}
              <button className="text-btn" onClick={() => setShowProjectForm(true)}>Новый проект <ArrowRight size={15} /></button>
            </div>
            <div className="focus-card"><div className="focus-icon"><Sparkles size={18} /></div><div className="eyebrow">МАЛЕНЬКИЙ СОВЕТ</div><h3>Делай важное<br />по одному шагу.</h3><p>Разбей большую цель на понятные задачи — так проще поддерживать темп.</p><div className="focus-decoration">✳</div></div>
          </div>
        </section>
      </div>
    </main>
    {showProjectForm && <div className="modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) setShowProjectForm(false); }}><form className="modal" onSubmit={createProject}><div className="modal-title"><div><div className="eyebrow">WORKSPACE</div><h2>Новый проект</h2></div><button type="button" className="icon-btn" onClick={() => setShowProjectForm(false)}><X size={18} /></button></div><label>Название<input value={projectForm.name} onChange={e => setProjectForm({ ...projectForm, name: e.target.value })} required maxLength={160} placeholder="Например, Учёба" /></label><label>Описание<textarea value={projectForm.description} onChange={e => setProjectForm({ ...projectForm, description: e.target.value })} rows={3} placeholder="Для чего этот проект?" /></label><button className="primary-btn full"><Plus size={16} /> Создать проект</button></form></div>}
    {showTaskForm && <div className="modal-backdrop" onMouseDown={e => { if (e.target === e.currentTarget) setShowTaskForm(false); }}><form className="modal" onSubmit={createTask}><div className="modal-title"><div><div className="eyebrow">TASK MANAGEMENT</div><h2>Новая задача</h2></div><button type="button" className="icon-btn" onClick={() => setShowTaskForm(false)}><X size={18} /></button></div><label>Название задачи<input value={taskForm.title} onChange={e => setTaskForm({ ...taskForm, title: e.target.value })} required maxLength={200} placeholder="Что нужно сделать?" /></label><label>Описание<textarea value={taskForm.description} onChange={e => setTaskForm({ ...taskForm, description: e.target.value })} rows={3} placeholder="Дополнительные детали" /></label><div className="form-row"><label>Проект<select required value={taskForm.project_id} onChange={e => setTaskForm({ ...taskForm, project_id: e.target.value })}><option value="">Выбери проект</option>{projects.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}</select></label><label>Приоритет<select value={taskForm.priority} onChange={e => setTaskForm({ ...taskForm, priority: e.target.value })}><option value="low">Низкий</option><option value="medium">Средний</option><option value="high">Высокий</option></select></label></div>{!projects.length && <div className="error">Сначала создай проект.</div>}<button className="primary-btn full" disabled={!projects.length}><Plus size={16} /> Создать задачу</button></form></div>}
    {selected && <div className="detail-drawer"><div className="drawer-head"><span className="eyebrow">TASK #{selected.id}</span><button className="icon-btn" onClick={() => setSelectedTask(null)}><X size={18} /></button></div><h2>{selected.title}</h2><p className="muted">{selected.description || "У этой задачи пока нет описания."}</p><div className="drawer-info"><span>Статус</span><strong>{selected.status === "completed" ? "Завершена" : selected.status === "in_progress" ? "В работе" : "Новая"}</strong><span>Приоритет</span><strong>{selected.priority}</strong><span>Проект</span><strong>{projects.find(p => p.id === selected.project_id)?.name || "—"}</strong></div><div className="comments-title"><h3>Комментарии</h3><span>{comments.length}</span></div><div className="comment-list">{comments.map(comment => <div className="comment" key={comment.id}><div className="avatar tiny">U</div><div><p>{comment.text}</p><small>{new Date(comment.created_at).toLocaleString("ru-RU")}</small></div></div>)}{!comments.length && <p className="muted no-comments">Пока нет комментариев.</p>}</div><form className="comment-form" onSubmit={addComment}><textarea value={commentText} onChange={e => setCommentText(e.target.value)} placeholder="Написать комментарий…" rows={2} required /><button className="primary-btn full"><ArrowRight size={15} /> Отправить</button></form></div>}
  </div>;
}
