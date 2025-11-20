// --- STATE MANAGEMENT ---
let currentLesson = null;
let currentSceneIndex = 0;

const views = {
  home: document.getElementById("view-home"),
  lesson: document.getElementById("view-lesson"),
};

const ui = {
  navbar: document.getElementById("navbar"),
  header: document.getElementById("header"),
  charAvatar: document.getElementById("charAvatar"),
  charName: document.getElementById("charName"),
  charDialogue: document.getElementById("charDialogue"),
  optionsContainer: document.getElementById("optionsContainer"),
  nextBtn: document.getElementById("nextBtn"),
  reactionImage: document.getElementById("reactionImage"),
};

// Tambahkan variable state untuk chat
let isChatOpen = false;

function toggleChat() {
    const panel = document.getElementById("aiTutorPanel");
    isChatOpen = !isChatOpen;
    if (isChatOpen) {
        panel.classList.remove("hidden");
    } else {
        panel.classList.add("hidden");
    }
}

function sendToAI() {
    const inputEl = document.getElementById("chatInput");
    const userText = inputEl.value;
    if (!userText) return;

    // 1. Tampilkan pesan user di UI
    addChatBubble(userText, "user");
    inputEl.value = "";

    // 2. Ambil Konteks Cerita Saat Ini
    // Ini kuncinya! Kita kirim dialog terakhir Sari ke AI
    const currentScene = currentLesson.scenes[currentSceneIndex];
    const context = `
        Karakter: ${currentScene.characterName}
        Mood: ${currentScene.characterMood}
        Dialog Karakter: "${currentScene.dialogue}"
        Jawaban Benar: ${currentScene.options.find(o => o.isCorrect).text}
    `;

    // 3. Tampilkan status "Typing..."
    const loadingId = addChatBubble("Sedang berpikir...", "ai", true);

    // 4. Kirim ke Backend Go
    fetch("/api/ask-tutor", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            context: context,
            userQuery: userText
        })
    })
    .then(res => res.json())
    .then(data => {
        // Hapus loading, ganti dengan jawaban asli
        document.getElementById(loadingId).remove();
        addChatBubble(data.reply, "ai");
    });
}

function addChatBubble(text, sender, isLoading = false) {
    const container = document.getElementById("chatHistory");
    const div = document.createElement("div");
    const id = "msg-" + Date.now();
    div.id = id;
    
    if (sender === "user") {
        div.className = "bg-blue-500 text-white p-3 rounded-l-xl rounded-tr-xl ml-auto max-w-[80%] text-sm";
    } else {
        div.className = "bg-slate-100 text-slate-800 p-3 rounded-r-xl rounded-tl-xl mr-auto max-w-[80%] text-sm border border-slate-200";
        if(isLoading) div.classList.add("italic", "text-slate-500");
    }
    
    div.innerText = text;
    container.appendChild(div);
    container.scrollTop = container.scrollHeight; // Auto scroll ke bawah
    return id;
}

// --- NAVIGATION ---
function show(viewName) {
  // Toggle visibility
  Object.values(views).forEach((el) => el.classList.add("hidden"));
  views[viewName].classList.remove("hidden");

  // Toggle Navbar/Header (Hide them inside lesson)
  if (viewName === "lesson") {
    ui.navbar.classList.add("hidden");
    ui.header.classList.add("hidden");
  } else {
    ui.navbar.classList.remove("hidden");
    ui.header.classList.remove("hidden");
  }
}

// --- API & DATA ---
function loadHome() {
  fetch("/api/state")
    .then((res) => res.json())
    .then((state) => {
      // Render Chips
      document.getElementById("chips").innerHTML = `
        <span class="bg-orange-100 text-orange-600 px-3 py-1 rounded-full text-sm font-bold">🔥 ${state.streakDays}</span>
        <span class="bg-blue-100 text-blue-600 px-3 py-1 rounded-full text-sm font-bold">💎 ${state.xp}</span>
      `;

      // Render Modules
      const container = document.getElementById("lessonPath");
      container.innerHTML = "";
      
      state.modules.forEach((mod) => {
        const btn = document.createElement("div");
        // Styling cards
        btn.className = `p-4 rounded-2xl border-b-4 active:border-b-0 active:translate-y-1 transition-all cursor-pointer flex items-center gap-4 ${
            mod.done ? "bg-slate-100 border-slate-200 text-slate-400" : "bg-white border-slate-200 shadow-sm hover:bg-slate-50"
        }`;
        
        btn.innerHTML = `
            <div class="w-14 h-14 rounded-full flex items-center justify-center text-2xl" style="background-color: ${mod.color}20;">${mod.icon}</div>
            <div>
                <div class="font-bold text-lg">${mod.title}</div>
                <div class="text-xs uppercase tracking-wide text-slate-500">${mod.done ? "Completed" : "Start Lesson"}</div>
            </div>
        `;
        
        // Click to start lesson
        btn.onclick = () => startLesson(mod.id);
        container.appendChild(btn);
      });
    });
}

function startLesson(id) {
    // In a real app, we fetch based on ID. Here we hardcode to lesson/1
    fetch("/api/lesson/1")
        .then(res => res.json())
        .then(lesson => {
            currentLesson = lesson;
            currentSceneIndex = 0;
            renderScene();
            show("lesson");
        });
}

// --- GAME LOOP UTAMA (Versi Baru) ---

function renderScene() {
    const scene = currentLesson.scenes[currentSceneIndex];
    
    // 1. Reset UI Dasar
    ui.nextBtn.classList.add("hidden");
    ui.charName.innerText = scene.characterName;
    
    // --- LOGIKA BARU: CEK APAKAH MOOD ITU GAMBAR ATAU EMOJI ---
    if (scene.characterMood.includes(".png")) {
        // JIKA GAMBAR (.png)
        ui.reactionImage.src = "/" + scene.characterMood; // Load gambar
        ui.reactionImage.classList.remove("opacity-0");   // Munculkan gambar
        ui.charAvatar.classList.add("opacity-0");         // Sembunyikan emoji 👩
    } else {
        // JIKA EMOJI (👋)
        ui.charAvatar.innerText = scene.characterMood;    // Set teks emoji
        ui.charAvatar.classList.remove("opacity-0");      // Munculkan emoji
        ui.reactionImage.classList.add("opacity-0");      // Sembunyikan gambar
        ui.reactionImage.src = ""; 
    }

    // 2. Efek Mengetik untuk Dialog Pertanyaan
    typeWriter(scene.dialogue);

    const choiceArea = document.getElementById("choiceArea");
    const inputArea = document.getElementById("inputArea");
    const textInput = document.getElementById("userTextInput");

    // 3. LOGIKA SWITCH: Apakah ini Choice atau Text?
    if (scene.inputType === "text") {
        choiceArea.classList.add("hidden");
        inputArea.classList.remove("hidden");
        textInput.value = ""; 
        textInput.focus();    
    } else {
        inputArea.classList.add("hidden");
        choiceArea.classList.remove("hidden");
        choiceArea.innerHTML = "";

        scene.options.forEach(opt => {
            const btn = document.createElement("button");
            btn.className = "w-full text-left p-4 rounded-xl border-2 border-slate-100 bg-slate-50 font-semibold text-slate-700 hover:border-blue-300 hover:bg-blue-50 transition-all active:scale-[0.98]";
            btn.innerText = opt.text;
            btn.onclick = () => handleChoiceAnswer(opt);
            choiceArea.appendChild(btn);
        });
    }
}

// --- HANDLER 1: JIKA USER KLIK PILIHAN GANDA ---
function handleChoiceAnswer(option) {
    // Disable semua tombol supaya tidak diklik 2x
    const buttons = document.getElementById("choiceArea").querySelectorAll("button");
    buttons.forEach(b => {
        b.disabled = true;
        if (b.innerText !== option.text) b.classList.add("opacity-40");
    });

    // Langsung proses hasilnya (karena data benar/salah sudah ada di JSON)
    processResult(option.isCorrect, option.reaction, option.text, option.reactionImage);
}

// --- HANDLER 2: JIKA USER KIRIM TEKS (Rendang/Sabun) ---
function submitTextAnswer() {
    const inputField = document.getElementById("userTextInput");
    const answer = inputField.value.trim();
    
    if (!answer) return; // Jangan kirim kalau kosong

    // Tampilkan status loading di dialog karakter
    ui.charDialogue.innerText = "(Sedang memproses jawaban...)";

    // Kirim ke Backend Go (/api/check-text)
    fetch("/api/check-text", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            sceneIndex: currentSceneIndex,
            answer: answer
        })
    })
    .then(res => res.json())
    .then(data => {
        // Sembunyikan input area supaya user tidak spam kirim
        document.getElementById("inputArea").classList.add("hidden");
        
        // Gabungkan Reaksi Karakter + Penjelasan Sensei (Jepang)
        // data.reaction = "Eh rendang?..."
        // data.feedback = "Itu makanan woy..."
        const fullReaction = `${data.reaction}\n\n(🇯🇵 Sensei: ${data.feedback})`;
        
        // Proses hasilnya
        processResult(data.isCorrect, fullReaction, answer, data.reactionImage);
    })
    .catch(err => {
        console.error(err);
        alert("Gagal menghubungi server.");
    });
}

// --- FUNGSI UMUM: MEMPROSES HASIL (BENAR/SALAH) ---
// Fungsi ini dipakai baik oleh Choice maupun Text supaya codingnya rapi
function processResult(isCorrect, reactionText, userResponse, reactionImageUrl) {
    // 1. Tampilkan Reaksi (Animasi Pop-in)
    ui.charDialogue.innerText = reactionText;
    ui.charDialogue.classList.add("pop-in");
    setTimeout(() => ui.charDialogue.classList.remove("pop-in"), 300);

    if (reactionImageUrl) {
        ui.reactionImage.src = "/" + reactionImageUrl; 
        ui.reactionImage.classList.remove("opacity-0"); // Munculkan Gambar PNG
        ui.charAvatar.classList.add("opacity-0");       // Sembunyikan Emoji
    } else {
        ui.reactionImage.classList.add("opacity-0");
        ui.reactionImage.src = "";
        ui.charAvatar.classList.remove("opacity-0");    // Munculkan Emoji kalau tidak ada gambar
    }

    // 2. Cek Benar atau Salah
    if (isCorrect) {
        // --- BENAR ---
        ui.charAvatar.innerText = "✨"; // Ekspresi Senang
        
        ui.nextBtn.innerText = "Continue";
        ui.nextBtn.className = "w-full py-4 bg-green-500 hover:bg-green-600 text-white font-bold rounded-xl shadow-lg shadow-green-500/30 transition-all";
        
        // Lanjut ke scene berikutnya
        ui.nextBtn.onclick = nextScene;

    } else {
        // --- SALAH ---
        ui.charAvatar.innerText = "😅"; // Ekspresi Bingung/Kaget
        
        ui.nextBtn.innerText = "Try Again";
        ui.nextBtn.className = "w-full py-4 bg-slate-200 text-slate-600 font-bold rounded-xl hover:bg-slate-300 transition-all";
        
        // Reset scene ini (Ulangi pertanyaan)
        ui.nextBtn.onclick = () => {
            renderScene();
        };
    }
    
    // Munculkan tombol Continue/Try Again
    ui.nextBtn.classList.remove("hidden");
}

function nextScene() {
    currentSceneIndex++;
    if (currentSceneIndex < currentLesson.scenes.length) {
        renderScene();
    } else {
        alert("Lesson Complete!");
        show("home");
    }
}

// Helper: Simple Text animation
function typeWriter(text) {
    ui.charDialogue.innerText = "";
    let i = 0;
    const speed = 20; 
    function type() {
        if (i < text.length) {
            ui.charDialogue.innerHTML += text.charAt(i);
            i++;
            setTimeout(type, speed);
        }
    }
    type();
}

// Init
loadHome();
show("home");