// ============================================================
// Defter — banka-projesi frontend
// Go REST API'ye (main.go) doğrudan fetch ile bağlanır.
// ============================================================

const STORAGE_KEY_URL = "defter_api_base";
const STORAGE_KEY_ID  = "defter_last_id";

let API_BASE = localStorage.getItem(STORAGE_KEY_URL) || "http://localhost:8080/api";

// Oturum durumu (sayfa yenilenince kaybolur — PIN hiçbir yerde saklanmaz)
let session = { id: null, pin: null, isim: null, bakiye: 0 };

// ---------- yardımcılar ----------
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

function tl(amount){
  return Number(amount).toLocaleString("tr-TR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function showToast(message, type = "info"){
  const stack = $("#toast-stack");
  const el = document.createElement("div");
  el.className = `toast ${type}`;
  el.textContent = message;
  stack.appendChild(el);
  setTimeout(() => {
    el.style.transition = "opacity .3s ease";
    el.style.opacity = "0";
    setTimeout(() => el.remove(), 300);
  }, 3200);
}

function setBusy(form, busy){
  const btn = form.querySelector("button[type=submit]");
  if (!btn) return;
  btn.disabled = busy;
  btn.querySelector(".btn-spinner").hidden = !busy;
}

// Go backend'ine POST atan tek merkezi fonksiyon.
async function api(path, body){
  let res;
  try{
    res = await fetch(`${API_BASE}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body ?? {}),
    });
  } catch (err){
    throw new Error(
      `Sunucuya ulaşılamadı (${API_BASE}). Go sunucusu çalışıyor mu ve CORS ayarlı mı? (⚙ ayarlar ikonundan adresi kontrol et)`
    );
  }

  let json;
  try{ json = await res.json(); }
  catch{ throw new Error("Sunucudan geçersiz yanıt geldi."); }

  if (!res.ok || json.success === false){
    throw new Error(json.message || `İstek başarısız (HTTP ${res.status})`);
  }
  return json.data;
}

// ---------- görünüm geçişleri ----------
function showDashboard(){
  $("#view-auth").hidden = true;
  $("#view-dashboard").hidden = false;
  $("#session-chip").hidden = false;

  $("#session-name").textContent = session.isim;
  $("#session-id").textContent = `#${String(session.id).padStart(4, "0")}`;
  $("#passbook-id").textContent = `#${String(session.id).padStart(4, "0")}`;
  $("#passbook-name").textContent = session.isim;
  $("#stamp-initial").textContent = (session.isim || "D").trim().charAt(0).toUpperCase();
  renderBalance(session.bakiye);
  refreshDashboardData();
}

function showAuth(){
  $("#view-dashboard").hidden = true;
  $("#view-auth").hidden = false;
  $("#session-chip").hidden = true;
  $("#asset-list").innerHTML = "";
}

function renderBalance(amount){
  session.bakiye = amount;
  $("#balance-amount").textContent = tl(amount);
}

function renderAssets(assets){
  const list = $("#asset-list");
  if (!assets || assets.length === 0){
    list.innerHTML = '<li class="asset-empty">Henüz döviz varlığınız yok.</li>';
    return;
  }

  list.innerHTML = assets.map((asset) => `
    <li class="asset-item">
      <div class="asset-main">
        <span class="asset-code">${asset.doviz_kodu}</span>
        <span class="asset-meta">Mevcut miktar</span>
      </div>
      <span class="asset-amount">${tl(asset.miktar)} ${asset.doviz_kodu}</span>
    </li>
  `).join("");
}

async function refreshDashboardData(){
  if (!session.id) return;

  try{
    const balanceData = await api("/bakiye", { id: session.id, pin: session.pin });
    renderBalance(balanceData.bakiye);
  } catch (err){
    showToast(err.message, "error");
  }

  try{
    const assets = await api("/varliklar", { id: session.id, pin: session.pin });
    renderAssets(assets);
  } catch (err){
    showToast(err.message, "error");
  }
}

// ---------- giriş ----------
$("#form-login").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;
  const id = Number($("#login-id").value);
  const pin = $("#login-pin").value;

  setBusy(form, true);
  try{
    const data = await api("/bakiye", { id, pin });
    session = { id: data.id, pin, isim: data.isim, bakiye: data.bakiye };
    localStorage.setItem(STORAGE_KEY_ID, String(id));
    showToast(`Hoş geldin, ${data.isim}.`, "success");
    showDashboard();
    form.reset();
  } catch(err){
    showToast(err.message, "error");
  } finally{
    setBusy(form, false);
  }
});

// ---------- hesap açma ----------
$("#form-create").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;
  const isim = $("#create-isim").value.trim();
  const bakiye = Number($("#create-bakiye").value || 0);
  const pin = $("#create-pin").value || "1234";

  setBusy(form, true);
  try{
    const data = await api("/hesap-ac", { isim, bakiye, pin });
    showToast(`Hesap oluşturuldu — Hesap No #${data.id}. Şimdi bu numarayla giriş yapabilirsin.`, "success");
    $("#login-id").value = data.id;
    form.reset();
  } catch(err){
    showToast(err.message, "error");
  } finally{
    setBusy(form, false);
  }
});

// ---------- çıkış / yenile ----------
$("#btn-logout").addEventListener("click", () => {
  session = { id: null, pin: null, isim: null, bakiye: 0 };
  showAuth();
});

$("#btn-refresh").addEventListener("click", async () => {
  const btn = $("#btn-refresh");
  btn.disabled = true;
  try{
    await refreshDashboardData();
    showToast("Veriler güncellendi.", "success");
  } catch(err){
    showToast(err.message, "error");
  } finally{
    btn.disabled = false;
  }
});

// ---------- sekmeler ----------
function switchTab(name){
  $$(".tab").forEach((t) => { t.classList.remove("active"); t.setAttribute("aria-selected", "false"); });
  const activeTab = document.querySelector(`.tab[data-tab="${name}"]`);
  if (activeTab){
    activeTab.classList.add("active");
    activeTab.setAttribute("aria-selected", "true");
  }

  $$(".panel").forEach((p) => { p.hidden = p.dataset.panel !== name; });

  if (name === "gecmis") loadHistory();
  if (name === "doviz") refreshDashboardData();
}

$$(".tab").forEach((tab) => {
  tab.addEventListener("click", () => switchTab(tab.dataset.tab));
});

$$(".quick-action").forEach((btn) => {
  btn.addEventListener("click", () => switchTab(btn.dataset.quickTab));
});

// ---------- para yatırma ----------
$("#form-yatir").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;
  const miktar = Number($("#yatir-miktar").value);

  setBusy(form, true);
  try{
    await api("/para-yatir", { id: session.id, miktar });
    await refreshDashboardData();
    showToast(`₺${tl(miktar)} yatırıldı.`, "success");
    form.reset();
  } catch(err){
    showToast(err.message, "error");
  } finally{
    setBusy(form, false);
  }
});

// ---------- para gönderme ----------
$("#form-gonder").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;
  const alici_id = Number($("#gonder-alici").value);
  const miktar = Number($("#gonder-miktar").value);
  const pin = $("#gonder-pin").value;

  setBusy(form, true);
  try{
    await api("/para-gonder", { gonderen_id: session.id, alici_id, miktar, pin });
    await refreshDashboardData();
    showToast(`₺${tl(miktar)} → Hesap #${alici_id} gönderildi.`, "success");
    form.reset();
  } catch(err){
    showToast(err.message, "error");
  } finally{
    setBusy(form, false);
  }
});

// ---------- döviz işlemleri ----------
$("#doviz-islem").addEventListener("change", () => {
  const label = $("#doviz-submit-label");
  label.textContent = $("#doviz-islem").value === "al" ? "Döviz Al" : "Döviz Sat";
});

$("#form-doviz").addEventListener("submit", async (e) => {
  e.preventDefault();
  const form = e.target;
  const islem = $("#doviz-islem").value;
  const kod = $("#doviz-kod").value.toUpperCase();
  const miktar = Number($("#doviz-miktar").value);
  const pin = $("#doviz-pin").value || session.pin;

  setBusy(form, true);
  try{
    await api(islem === "al" ? "/doviz-al" : "/doviz-sat", { hesap_id: session.id, doviz_kodu: kod, miktar, pin });
    await refreshDashboardData();
    showToast(islem === "al" ? `${kod} alımı tamamlandı.` : `${kod} satışı tamamlandı.`, "success");
    form.reset();
  } catch(err){
    showToast(err.message, "error");
  } finally{
    setBusy(form, false);
  }
});

// ---------- işlem geçmişi ----------
async function loadHistory(){
  const list = $("#ledger-list");
  const empty = $("#ledger-empty");
  list.innerHTML = "";

  try{
    const data = await api("/gecmis", { id: session.id, pin: session.pin });
    if (!data || data.length === 0){
      empty.hidden = false;
      return;
    }
    empty.hidden = true;

    data.forEach((islem) => {
      let direction = "credit";
      let label = "Bilinmeyen işlem";

      if (islem.islem_tipi === "YATIRMA") {
        label = "Para Yatırma";
        direction = "credit";
      } else if (islem.islem_tipi === "TRANSFER") {
        const isOutgoing = islem.gonderen_id === session.id;
        direction = isOutgoing ? "debit" : "credit";
        label = isOutgoing
          ? `Hesap #${islem.alici_id}'e transfer`
          : `Hesap #${islem.gonderen_id}'den transfer`;
      } else if (islem.islem_tipi === "ALIM") {
        label = `Döviz Alım — ${islem.doviz_kodu || "?"}`;
        direction = "debit";
      } else if (islem.islem_tipi === "SATIM") {
        label = `Döviz Satış — ${islem.doviz_kodu || "?"}`;
        direction = "credit";
      } else {
        label = islem.islem_tipi;
      }

      const li = document.createElement("li");
      li.className = "ledger-row";
      li.innerHTML = `
        <span class="ledger-dot ${direction}"></span>
        <div class="ledger-body">
          <span class="ledger-type">${label}</span>
          <span class="ledger-meta">${formatDate(islem.tarih)}</span>
        </div>
        <span class="ledger-amount ${direction}">${direction === "debit" ? "−" : "+"}₺${tl(islem.miktar_tl)}</span>
      `;
      list.appendChild(li);
    });
  } catch(err){
    showToast(err.message, "error");
  }
}

function formatDate(raw){
  if (!raw) return "";
  const d = new Date(raw.replace(" ", "T"));
  if (isNaN(d)) return raw;
  return d.toLocaleString("tr-TR", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" });
}

// ---------- sunucu ayarları ----------
$("#btn-settings").addEventListener("click", () => {
  $("#settings-url").value = API_BASE;
  $("#settings-backdrop").hidden = false;
});
$("#settings-cancel").addEventListener("click", () => { $("#settings-backdrop").hidden = true; });
$("#settings-backdrop").addEventListener("click", (e) => {
  if (e.target.id === "settings-backdrop") $("#settings-backdrop").hidden = true;
});
$("#settings-save").addEventListener("click", () => {
  const val = $("#settings-url").value.trim().replace(/\/$/, "");
  if (val){
    API_BASE = val;
    localStorage.setItem(STORAGE_KEY_URL, val);
    showToast("Sunucu adresi güncellendi.", "success");
  }
  $("#settings-backdrop").hidden = true;
});

// ---------- başlangıç ----------
(function init(){
  const lastId = localStorage.getItem(STORAGE_KEY_ID);
  if (lastId) $("#login-id").value = lastId;
})();
