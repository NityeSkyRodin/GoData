let raceIndex = [];
let championship = [];
let teamChampionship = [];
let currentRace = null;
let currentSession = 'race';
let standingsMode = 'drivers';

document.addEventListener('DOMContentLoaded', async () => {
    initEventListeners();
    try {
        await loadChampionship();
        await loadRaceIndex();
        renderChampionship();
        setupSidebar();
    } catch (error) {
        console.error(error);
    }
});

function initEventListeners() {
    document.getElementById('nav-overview').addEventListener('click', showOverview);
    document.getElementById('mobile-menu-btn').addEventListener('click', () => {
        document.getElementById('sidebar').classList.toggle('-translate-x-full');
    });
    document.getElementById('toggle-drivers').addEventListener('click', () => setStandingsMode('drivers'));
    document.getElementById('toggle-teams').addEventListener('click', () => setStandingsMode('teams'));
}

async function loadChampionship() {
    const response = await fetch('./data/championship.json');
    if (!response.ok) {
        throw new Error('championship.json not found');
    }
    championship = await response.json();
    calculateTeamStandings();
}

function calculateTeamStandings() {
    const teamMap = new Map();
    championship.forEach(driver => {
        const team = driver.team_name || 'Unknown Team';
        const points = parseInt(driver.points) || 0;
        teamMap.set(team, (teamMap.get(team) || 0) + points);
    });
    teamChampionship = Array.from(teamMap, ([team_name, points]) => ({ team_name, points }));
    teamChampionship.sort((a, b) => b.points - a.points);
}

async function loadRaceIndex() {
    const response = await fetch('./data/races/index.json');
    if (!response.ok) {
        throw new Error('races/index.json not found');
    }
    raceIndex = await response.json();
}

function setupSidebar() {
    const raceList = document.getElementById('race-list');
    raceList.innerHTML = '';
    raceIndex.forEach((race) => {
        const button = document.createElement('button');
        button.className = 'w-full flex items-center gap-3 px-3 py-3 slanted bg-[#151821] text-slate-400 hover:text-white hover:bg-[#232734] transition-all group';
        button.innerHTML = `
            <span class="text-[#0090D0] font-black italic">/</span>
            <span class="font-bold tracking-wider text-xs uppercase">${race.track_name}</span>
        `;
        button.addEventListener('click', () => loadRace(race));
        raceList.appendChild(button);
    });
}

function showOverview() {
    document.getElementById('view-overview').classList.remove('hidden');
    document.getElementById('view-race').classList.add('hidden');
    updateActiveNav('nav-overview');
}

// ---------- Shared card templates ----------

function twoDigits(number) {
    return String(number).padStart(2, '0');
}

function statCard(label, value, { sub = '', accent = false, index = 0 } = {}) {
    // Numbers get a big font, names a smaller uppercase one so they still fit
    const isNumber = typeof value === 'number' || /^[\d/]+$/.test(String(value));
    const valueClass = isNumber ? 'text-4xl' : 'text-xl uppercase tracking-wide truncate';
    return `
        <div class="relative overflow-hidden bg-[#10141C] border border-[#232734] corner-cut p-5 min-h-[104px]">
            <div class="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-500">${label}</div>
            <div class="font-header font-black mt-2 ${valueClass} ${accent ? 'text-[#0090D0]' : 'text-white'}">${value}</div>
            ${sub ? `<div class="text-[10px] font-bold uppercase tracking-[0.15em] text-[#0090D0] mt-1 truncate">${sub}</div>` : ''}
            <div class="watermark">${twoDigits(index + 1)}</div>
        </div>
    `;
}

function podiumCard({ position, name, sub, value, watermark }) {
    const first = position === 1;
    return `
        <div class="relative overflow-hidden corner-cut p-5 min-h-[120px] ${first ? 'bg-[#0090D0]' : 'bg-[#10141C] border border-[#232734]'}">
            <div class="flex justify-between items-start gap-3">
                <span class="font-header text-3xl font-black italic ${first ? 'text-white/50' : 'text-slate-500'}">P${position}</span>
                <span class="font-header text-xl font-black text-white whitespace-nowrap">${value}</span>
            </div>
            <div class="relative z-10 font-header text-2xl font-black uppercase text-white mt-1 truncate">${name}</div>
            ${sub ? `<div class="relative z-10 text-[10px] font-bold uppercase tracking-[0.2em] mt-1 truncate ${first ? 'text-white/80' : 'text-[#0090D0]'}">${sub}</div>` : ''}
            <div class="watermark ${first ? 'watermark-on-blue' : ''}">${watermark ?? ''}</div>
        </div>
    `;
}

// ---------- Overview ----------

function renderChampionship() {
    renderStats();
    renderStandings();
    renderCalendar();
}

function renderStats() {
    const leader = championship[0];
    const latest = raceIndex[raceIndex.length - 1];
    document.getElementById('stats-cards').innerHTML = [
        statCard('Races', raceIndex.length, { accent: true, index: 0 }),
        statCard('Drivers', championship.length, { index: 1 }),
        statCard('Leader Pts', leader?.points ?? 0, { sub: leader?.driver_name, index: 2 }),
        statCard('Latest GP', latest?.track_name ?? '--', { index: 3 })
    ].join('');
}

function setStandingsMode(mode) {
    standingsMode = mode;
    const isDrivers = mode === 'drivers';

    const btnDrivers = document.getElementById('toggle-drivers');
    btnDrivers.classList.toggle('bg-[#0090D0]', isDrivers);
    btnDrivers.classList.toggle('text-white', isDrivers);
    btnDrivers.classList.toggle('text-slate-500', !isDrivers);

    const btnTeams = document.getElementById('toggle-teams');
    btnTeams.classList.toggle('bg-[#0090D0]', !isDrivers);
    btnTeams.classList.toggle('text-white', !isDrivers);
    btnTeams.classList.toggle('text-slate-500', isDrivers);

    renderStandings();
}

function renderStandings() {
    const isDrivers = standingsMode === 'drivers';
    const data = isDrivers ? championship : teamChampionship;

    // Top 3 as cards
    document.getElementById('standings-podium').innerHTML = data.slice(0, 3).map((item, index) => podiumCard({
        position: index + 1,
        name: isDrivers ? item.driver_name : item.team_name,
        sub: isDrivers ? item.team_name : '',
        value: `${item.points} PTS`,
        watermark: isDrivers ? item.driver_number : index + 1
    })).join('');

    // Everyone from P4 down in the table
    document.getElementById('standings-name-header').textContent = isDrivers ? 'Driver' : 'Team';
    document.getElementById('standings-number-header').classList.toggle('hidden', !isDrivers);
    document.getElementById('standings-team-header').classList.toggle('hidden', !isDrivers);

    const rest = data.slice(3);
    document.getElementById('standings-table-wrapper').classList.toggle('hidden', rest.length === 0);
    document.getElementById('standings-table-body').innerHTML = rest.map((item, index) => `
        <tr class="hover:bg-[#151821] transition-colors border-b border-[#232734]/60 last:border-0">
            <td class="py-4 px-5 font-header font-black text-xl text-slate-500">${index + 4}</td>
            <td class="py-4 px-5 font-header font-bold text-lg text-slate-500 ${!isDrivers ? 'hidden' : ''}">${item.driver_number ?? '-'}</td>
            <td class="py-4 px-5 font-header font-bold text-lg text-white uppercase tracking-wide">${isDrivers ? item.driver_name : item.team_name}</td>
            <td class="py-4 px-5 text-xs font-bold uppercase tracking-[0.15em] text-slate-500 ${!isDrivers ? 'hidden' : ''}">${item.team_name}</td>
            <td class="py-4 px-5 text-right font-header font-black text-xl text-[#0090D0]">${item.points}</td>
        </tr>
    `).join('');
}

function renderCalendar() {
    const container = document.getElementById('race-calendar');
    container.innerHTML = raceIndex.map((race, index) => `
        <button data-index="${index}" class="calendar-card group relative overflow-hidden text-left bg-[#10141C] border border-[#232734] hover:border-[#0090D0] corner-cut p-5 min-h-[104px] transition-colors">
            <div class="text-[10px] font-bold uppercase tracking-[0.2em] text-[#0090D0]">Round ${twoDigits(index + 1)}</div>
            <div class="flex items-center justify-between gap-3 mt-4">
                <span class="font-header text-xl font-black uppercase tracking-wide text-white truncate">${race.track_name}</span>
                <span class="text-2xl text-slate-600 group-hover:text-[#0090D0] group-hover:translate-x-1 transition-all">&rarr;</span>
            </div>
        </button>
    `).join('');

    container.querySelectorAll('.calendar-card').forEach(button => {
        button.addEventListener('click', () => loadRace(raceIndex[button.dataset.index]));
    });
}

// ---------- Race weekend ----------

async function loadRace(race) {
    try {
        const response = await fetch(`./data/races/${race.file}`);
        if (!response.ok) {
            throw new Error(`Could not load ${race.file}`);
        }
        currentRace = await response.json();
        renderRace(currentRace, race);
    } catch (error) {
        console.error(error);
    }
}

function renderRace(race, raceIndexEntry) {
    document.getElementById('view-overview').classList.add('hidden');
    document.getElementById('view-race').classList.remove('hidden');
    document.getElementById('sidebar').classList.add('-translate-x-full');
    updateActiveNav(null);
    window.scrollTo(0, 0);

    const round = raceIndex.indexOf(raceIndexEntry) + 1;
    document.getElementById('race-title').textContent = race.track_name;
    document.getElementById('race-subtitle').textContent = `Round ${twoDigits(round)} · Season 2026`;

    renderSessionTabs(race);
    selectSession(getFirstAvailableSession(race));
}

function renderSessionTabs(race) {
    const container = document.getElementById('session-tabs');
    const sessions = [
        { key: 'qualifying', label: 'Qualifying' },
        { key: 'sprint', label: 'Sprint' },
        { key: 'race', label: 'Feature' }
    ];

    container.innerHTML = sessions
        .filter(session => race[session.key]?.length)
        .map(session => `
            <button data-session="${session.key}" class="session-tab px-5 py-2.5 corner-cut font-header text-base font-bold uppercase tracking-[0.15em] whitespace-nowrap text-slate-500 hover:text-white transition-colors">
                ${session.label}
            </button>
        `).join('');

    container.querySelectorAll('.session-tab').forEach(button => {
        button.addEventListener('click', () => selectSession(button.dataset.session));
    });
}

function getFirstAvailableSession(race) {
    if (race.race?.length) return 'race';
    if (race.sprint?.length) return 'sprint';
    return 'qualifying';
}

function selectSession(session) {
    currentSession = session;
    const results = currentRace?.[session] ?? [];

    document.querySelectorAll('.session-tab').forEach(button => {
        const active = button.dataset.session === session;
        button.classList.toggle('bg-[#0090D0]', active);
        button.classList.toggle('text-white', active);
        button.classList.toggle('text-slate-500', !active);
    });

    renderSession(results, session);
}

function renderSession(results, session) {
    const sessionNames = { qualifying: 'Qualifying', sprint: 'Sprint Race', race: 'Feature Race' };
    document.getElementById('session-title').textContent = sessionNames[session] ?? session;

    renderRaceStats(results);

    const sorted = [...results].sort((a, b) => a.position - b.position);

    // Qualifying has no points worth showing, so the podium shows the lap time instead
    document.getElementById('race-podium').innerHTML = sorted.slice(0, 3).map(driver => podiumCard({
        position: driver.position,
        name: driver.driver_name,
        sub: driver.team_name,
        value: session === 'qualifying' ? (driver.best_lap_time ?? '-') : `${driver.points ?? 0} PTS`,
        watermark: driver.race_number
    })).join('');

    const rest = sorted.slice(3);
    document.getElementById('race-table-wrapper').classList.toggle('hidden', rest.length === 0);
    document.getElementById('race-table-body').innerHTML = rest.map(driver => {
        const finished = String(driver.status).toLowerCase() === 'finished';
        return `
            <tr class="hover:bg-[#151821] transition-colors border-b border-[#232734]/60 last:border-0">
                <td class="py-4 px-4 font-header font-black text-xl text-slate-500">${driver.position}</td>
                <td class="py-4 px-4 font-header font-bold text-lg text-slate-500">${driver.race_number ?? '-'}</td>
                <td class="py-4 px-4 font-header font-bold text-lg text-white uppercase tracking-wide">${driver.driver_name}</td>
                <td class="py-4 px-4 text-xs font-bold uppercase tracking-[0.15em] text-slate-500">${driver.team_name ?? '-'}</td>
                <td class="py-4 px-4 text-center font-header font-bold text-lg text-slate-300">${driver.num_laps ?? '-'}</td>
                <td class="py-4 px-4 font-header font-bold text-base text-slate-400">${driver.best_lap_time ?? '-'}</td>
                <td class="py-4 px-4 font-header font-bold text-base text-slate-400">${driver.total_race_time ?? '-'}</td>
                <td class="py-4 px-4 text-center">
                    <span class="inline-flex px-3 py-1 corner-cut text-[10px] font-black uppercase tracking-widest ${finished ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}">
                        ${driver.status ?? '-'}
                    </span>
                </td>
                <td class="py-4 px-4 text-right font-header font-black text-xl ${driver.points > 0 ? 'text-[#0090D0]' : 'text-slate-600'}">${driver.points ?? 0}</td>
            </tr>
        `;
    }).join('');
}

function parseLapTime(time) {
    // "1:37.562" -> 97.562 seconds, null when there is no valid time
    const match = /^(?:(\d+):)?(\d+(?:\.\d+)?)$/.exec(String(time ?? '').trim());
    if (!match) return null;
    const seconds = (parseInt(match[1] ?? '0') * 60) + parseFloat(match[2]);
    return seconds > 0 ? seconds : null;
}

function renderRaceStats(results) {
    const winner = results.find(driver => driver.position === 1);
    const pole = currentRace?.qualifying?.find(driver => driver.position === 1);

    const fastest = results
        .filter(driver => parseLapTime(driver.best_lap_time) !== null)
        .sort((a, b) => parseLapTime(a.best_lap_time) - parseLapTime(b.best_lap_time))[0];

    const finishers = results.filter(driver => String(driver.status).toLowerCase() === 'finished').length;
    const dnfs = results.length - finishers;

    document.getElementById('race-stats-cards').innerHTML = [
        statCard('Winner', winner?.driver_name ?? '--', { sub: winner?.team_name, index: 0 }),
        statCard('Fastest Lap', fastest?.driver_name ?? '--', { sub: fastest?.best_lap_time, index: 1 }),
        statCard('Pole Position', pole?.driver_name ?? '--', { sub: pole?.best_lap_time, index: 2 }),
        statCard('Finished', `${finishers}/${results.length}`, { sub: `${dnfs} DNF`, accent: dnfs === 0, index: 3 })
    ].join('');
}

function updateActiveNav(activeId) {
    document.querySelectorAll('#sidebar button').forEach(button => {
        if (button.id === 'mobile-menu-btn') return;
        const active = button.id === activeId;
        button.classList.toggle('bg-[#0090D0]', active);
        button.classList.toggle('text-white', active);
        button.classList.toggle('shadow-lg', active);
        button.classList.toggle('shadow-[#0090D0]/20', active);
        button.classList.toggle('bg-[#151821]', !active);
        button.classList.toggle('text-slate-400', !active);
    });
}
