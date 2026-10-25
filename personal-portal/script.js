const menuButton = document.querySelector('.menu-toggle');
const nav = document.querySelector('.topbar nav');

menuButton.addEventListener('click', () => {
  const isOpen = nav.classList.toggle('open');
  menuButton.setAttribute('aria-expanded', String(isOpen));
});

nav.querySelectorAll('a').forEach((link) => {
  link.addEventListener('click', () => {
    nav.classList.remove('open');
    menuButton.setAttribute('aria-expanded', 'false');
  });
});

document.querySelectorAll('.chibi').forEach((figure) => {
  let animationTimer;

  figure.addEventListener('click', () => {
    window.clearTimeout(animationTimer);
    figure.classList.remove('is-cute');
    void figure.offsetWidth;
    figure.classList.add('is-cute');
    animationTimer = window.setTimeout(() => {
      figure.classList.remove('is-cute');
    }, 1400);
  });
});

const themeButton = document.querySelector('#theme-toggle');
const themeSymbol = themeButton.querySelector('.theme-symbol');
const themeState = themeButton.querySelector('.theme-state');
const themeModes = ['system', 'light', 'dark'];
const themeLabels = { system: '系统', light: '白天', dark: '黑夜' };
let themeMode;

try {
  themeMode = localStorage.getItem('abyssa1-theme-mode') || 'system';
} catch {
  themeMode = 'system';
}
if (!themeModes.includes(themeMode)) themeMode = 'system';

function applyTheme() {
  const prefersLight = window.matchMedia('(prefers-color-scheme: light)').matches;
  const resolvedTheme = themeMode === 'system' ? (prefersLight ? 'light' : 'dark') : themeMode;
  document.body.dataset.theme = resolvedTheme;
  themeSymbol.textContent = themeMode === 'system' ? '◐' : (themeMode === 'light' ? '☼' : '☾');
  themeState.textContent = themeLabels[themeMode];
  themeButton.setAttribute('aria-label', `外观模式：${themeLabels[themeMode]}，点击切换`);
  themeButton.title = `外观：${themeLabels[themeMode]}（点击切换）`;
}

themeButton.addEventListener('click', () => {
  const currentIndex = themeModes.indexOf(themeMode);
  themeMode = themeModes[(currentIndex + 1) % themeModes.length];
  try {
    localStorage.setItem('abyssa1-theme-mode', themeMode);
  } catch {
    // Theme switching still works when browser storage is unavailable.
  }
  applyTheme();
});

window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
  if (themeMode === 'system') applyTheme();
});
applyTheme();

const musicButton = document.querySelector('#music-toggle');
const backgroundAudio = document.querySelector('#background-audio');
const controlToast = document.querySelector('#control-toast');
const MUSIC_SOURCE = '';
let toastTimer;

function showControlToast(message) {
  controlToast.textContent = message;
  controlToast.classList.add('visible');
  window.clearTimeout(toastTimer);
  toastTimer = window.setTimeout(() => controlToast.classList.remove('visible'), 2400);
}

if (MUSIC_SOURCE) {
  backgroundAudio.src = MUSIC_SOURCE;
  musicButton.setAttribute('aria-label', '播放背景音乐');
  musicButton.dataset.hint = '播放背景音乐';
}

musicButton.addEventListener('click', async () => {
  if (!MUSIC_SOURCE) {
    showControlToast('等你选好歌曲，再把音乐接进来');
    musicButton.classList.remove('needs-track');
    void musicButton.offsetWidth;
    musicButton.classList.add('needs-track');
    return;
  }

  try {
    if (backgroundAudio.paused) {
      await backgroundAudio.play();
      musicButton.classList.add('is-playing');
      musicButton.setAttribute('aria-pressed', 'true');
      musicButton.setAttribute('aria-label', '暂停背景音乐');
      musicButton.dataset.hint = '正在播放 · 点击暂停';
    } else {
      backgroundAudio.pause();
      musicButton.classList.remove('is-playing');
      musicButton.setAttribute('aria-pressed', 'false');
      musicButton.setAttribute('aria-label', '播放背景音乐');
      musicButton.dataset.hint = '点击播放背景音乐';
    }
  } catch {
    showControlToast('音乐暂时无法播放，请检查音频地址');
  }
});

backgroundAudio.addEventListener('ended', () => {
  musicButton.classList.remove('is-playing');
  musicButton.setAttribute('aria-pressed', 'false');
});

const dailyThoughts = [
  '不必等到准备完美，先让第一步发生。',
  '慢一点没关系，别停在原地就好。',
  '今天解决一个小问题，也是在靠近想去的地方。',
  '把注意力放回能做的事，答案会慢慢清晰。',
  '每一次认真尝试，都在为下一次积累经验。',
  '先完成，再打磨；先开始，再调整。',
  '走得稳的人，也会走得很远。',
  '好奇心是起点，持续行动让它抵达远方。',
  '不和昨天的自己比较速度，只看看是否又前进了一点。',
  '允许计划改变，但别忘了最初想去的方向。',
  '把复杂的事拆小，今天只做好眼前这一块。',
  '休息不是退后，是为了带着力气继续。',
  '看不见进步的时候，积累也正在发生。',
  '勇气不一定很响亮，有时只是再试一次。',
  '认真做过的事，会留下比结果更长久的东西。',
  '别让一次失手替你决定下一次尝试。',
  '新的可能，常常藏在熟悉的事情之外。',
  '把想法写下来，它就有机会变成计划。',
  '无需一次走完全程，找到下一步就够了。',
  '耐心不是等待，而是继续做值得做的事。',
  '今天的专注，是送给未来自己的礼物。',
  '遇到难题时，换个角度也算向前。',
  '做出一点真实的东西，胜过反复等待灵感。',
  '给自己一点时间，也给目标一点坚持。',
  '把失败当作线索，下一次就多一分把握。',
  '方向可以边走边校准，脚步先迈出去。',
  '微小的好习惯，会在日子里长成底气。',
  '专注眼前的一件事，让它成为今天的进展。',
  '不需要每一天都耀眼，只要仍然在生长。',
  '愿你带着好奇出发，也带着新的答案回来。'
];

const quoteText = document.querySelector('#daily-quote-text');
if (quoteText) {
  const now = new Date();
  const localMidnight = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const dayNumber = Math.floor(Date.UTC(localMidnight.getFullYear(), localMidnight.getMonth(), localMidnight.getDate()) / 86400000);
  const quoteIndex = ((dayNumber % dailyThoughts.length) + dailyThoughts.length) % dailyThoughts.length;
  const quoteDate = document.querySelector('#daily-quote-date');
  const quoteCount = document.querySelector('#daily-quote-count');

  quoteText.textContent = dailyThoughts[quoteIndex];
  quoteDate.textContent = new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric', month: 'long', day: 'numeric'
  }).format(now);
  quoteCount.textContent = `DAY ${String(quoteIndex + 1).padStart(2, '0')} / ${dailyThoughts.length}`;
}
