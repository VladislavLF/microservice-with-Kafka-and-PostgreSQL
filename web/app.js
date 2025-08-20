document.getElementById('btn').addEventListener('click', async () => {
  const id = document.getElementById('oid').value.trim();
  const out = document.getElementById('result');
  out.textContent = 'Загрузка...';
  try {
    const resp = await fetch(`/order/${encodeURIComponent(id)}`);
    if (!resp.ok) {
      out.textContent = `Ошибка: ${resp.status}`;
      return;
    }
    const data = await resp.json();
    out.textContent = JSON.stringify(data, null, 2);
  } catch (e) {
    out.textContent = 'Ошибка сети';
  }
});
