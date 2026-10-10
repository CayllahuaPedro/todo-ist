const NEXT_STATUS = {
  pending: "in_progress",
  in_progress: "completed",
  completed: "pending",
  cancelled: "pending",
};

const STATUS_LABEL = {
  pending: "Not started",
  in_progress: "In progress",
  completed: "Done",
  cancelled: "Cancelled",
};

const list = document.getElementById("tasks");
const form = document.getElementById("new-task");
const input = document.getElementById("title");
const notice = document.getElementById("notice");
const summary = document.getElementById("summary");

let tasks = [];

async function request(path, options = {}) {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const message = (await res.text()).trim();
    throw new Error(message || `Request failed (${res.status})`);
  }
  return res.status === 204 ? null : res.json();
}

function descriptionOf(task) {
  return task.description?.Valid ? task.description.String : "";
}

function renderSummary() {
  if (tasks.length === 0) {
    summary.textContent = "";
    return;
  }
  const done = tasks.filter((t) => t.status === "completed").length;
  const open = tasks.filter((t) => t.status === "pending" || t.status === "in_progress").length;
  summary.textContent = `${open} open, ${done} done`;
}

function renderTask(task) {
  const li = document.createElement("li");
  li.className = "task";
  li.dataset.status = task.status;

  const mark = document.createElement("button");
  mark.type = "button";
  mark.className = "mark";
  mark.setAttribute(
    "aria-label",
    `${STATUS_LABEL[task.status]}. Mark as ${STATUS_LABEL[NEXT_STATUS[task.status]].toLowerCase()}`,
  );
  mark.addEventListener("click", () => advance(task));

  const body = document.createElement("div");
  body.className = "task-body";

  const title = document.createElement("p");
  title.className = "task-title";
  title.textContent = task.title;
  body.append(title);

  const details = descriptionOf(task);
  if (details || task.status === "in_progress") {
    const meta = document.createElement("p");
    meta.className = "task-meta";
    meta.textContent = [task.status === "in_progress" ? "In progress" : "", details]
      .filter(Boolean)
      .join(". ");
    body.append(meta);
  }

  const remove = document.createElement("button");
  remove.type = "button";
  remove.className = "remove";
  remove.textContent = "Delete";
  remove.setAttribute("aria-label", `Delete "${task.title}"`);
  remove.addEventListener("click", () => removeTask(task));

  li.append(mark, body, remove);
  return li;
}

function render() {
  list.replaceChildren();
  if (tasks.length === 0) {
    const empty = document.createElement("li");
    empty.className = "empty";
    empty.textContent = "Nothing on the list. Add your first task above.";
    list.append(empty);
  } else {
    list.append(...tasks.map(renderTask));
  }
  renderSummary();
}

function showError(err) {
  notice.textContent = err instanceof TypeError
    ? "Can't reach the server. Check that it's running on port 8080, then try again."
    : err.message;
}

async function load() {
  try {
    tasks = (await request("/todos")) ?? [];
    notice.textContent = "";
    render();
  } catch (err) {
    showError(err);
  }
}

async function advance(task) {
  try {
    await request(`/todos/${task.id}`, {
      method: "PATCH",
      body: JSON.stringify({
        title: task.title,
        description: descriptionOf(task),
        status: NEXT_STATUS[task.status],
      }),
    });
    await load();
  } catch (err) {
    showError(err);
  }
}

async function removeTask(task) {
  try {
    await request(`/todos/${task.id}`, { method: "DELETE" });
    await load();
  } catch (err) {
    showError(err);
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const title = input.value.trim();
  if (!title) return;

  const button = form.querySelector("button");
  button.disabled = true;
  try {
    await request("/todos", {
      method: "POST",
      body: JSON.stringify({ title, status: "pending" }),
    });
    input.value = "";
    await load();
  } catch (err) {
    showError(err);
  } finally {
    button.disabled = false;
    input.focus();
  }
});

load();
