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
    const response = await fetch('../../data/championship.json');
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
    const response = await fetch('../../data/races/index.json');
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

function renderChampionship() {
    renderStats();
    renderStandings();
}

function renderStats() {
    const leader = championship[0];
    const container = document.getElementById('stats-cards');
    container.innerHTML = `
        <div class="bg-[#151821] p-5 slanted border border-[#232734]">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Driver Leader</div>
            <div class="text-xl font-black text-white mt-1 font-header italic tracking-wide truncate">${leader?.driver_name ?? '--'}</div>
            <div class="text-sm text-[#0090D0] font-bold mt-1">${leader?.points ?? 0} PTS</div>
        </div>
        <div class="bg-[#151821] p-5 slanted border border-[#232734]">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Team Leader</div>
            <div class="text-xl font-black text-white mt-1 font-header italic tracking-wide truncate">${teamChampionship[0]?.team_name ?? '--'}</div>
            <div class="text-sm text-[#0090D0] font-bold mt-1">${teamChampionship[0]?.points ?? 0} PTS</div>
        </div>
        <div class="bg-[#151821] p-5 slanted border border-[#232734]">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Drivers</div>
            <div class="text-2xl font-black text-white mt-1 font-header italic">${championship.length}</div>
            <div class="text-xs text-slate-400 font-bold mt-1 uppercase">Classified</div>
        </div>
        <div class="bg-[#151821] p-5 slanted border border-[#232734]">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Races</div>
            <div class="text-2xl font-black text-white mt-1 font-header italic">${raceIndex.length}</div>
            <div class="text-xs text-slate-400 font-bold mt-1 uppercase">Completed</div>
        </div>
    `;
    document.getElementById('season-status').textContent = `${raceIndex.length} race${raceIndex.length === 1 ? '' : 's'} completed`;
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
    const tbody = document.getElementById('standings-table-body');
    const data = standingsMode === 'drivers' ? championship : teamChampionship;
    const isDrivers = standingsMode === 'drivers';
    
    document.getElementById('standings-name-header').textContent = isDrivers ? 'Driver' : 'Team';
    document.getElementById('standings-team-header').classList.toggle('hidden', !isDrivers);

    tbody.innerHTML = data.map((item, index) => `
        <tr class="hover:bg-[#232734]/50 transition-colors border-b border-[#232734]/50 group">
            <td class="py-4 px-4 font-black font-header text-2xl italic ${index === 0 ? 'text-[#0090D0]' : 'text-slate-600'}">
                ${index + 1}
            </td>
            <td class="py-4 px-4 font-bold text-white tracking-wide uppercase text-sm">
                ${isDrivers ? item.driver_name : item.team_name}
            </td>
            <td class="py-4 px-4 text-slate-400 text-xs font-bold uppercase tracking-wider ${!isDrivers ? 'hidden' : ''}">
                ${item.team_name}
            </td>
            <td class="py-4 px-4 text-right font-black font-header text-2xl italic text-white group-hover:text-[#0090D0] transition-colors">
                ${item.points}
            </td>
        </tr>
    `).join('');
}

async function loadRace(race) {
    try {
        const response = await fetch(`../../data/races/${race.file}`);
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

    document.getElementById('race-title').textContent = race.track_name;
    document.getElementById('race-subtitle').textContent = `Track ID ${race.track_id}`;

    renderSessionTabs(race);
    selectSession(getFirstAvailableSession(race));
}

function renderSessionTabs(race) {
    const container = document.getElementById('session-tabs');
    const sessions = [
        { key: 'qualifying', label: 'Qualifying' },
        { key: 'sprint', label: 'Sprint' },
        { key: 'race', label: 'Feature Race' }
    ];

    container.innerHTML = sessions
        .filter(session => race[session.key]?.length)
        .map(session => `
            <button data-session="${session.key}" class="session-tab px-6 py-2.5 slanted-tab text-xs font-bold uppercase tracking-widest bg-[#151821] text-slate-500 hover:text-white transition-all">
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
        button.classList.toggle('bg-[#151821]', !active);
        button.classList.toggle('text-slate-500', !active);
    });

    renderSession(results, session);
}

function renderSession(results, session) {
    const title = document.getElementById('session-title');
    title.textContent = `${session === 'race' ? 'Feature Race' : session.charAt(0).toUpperCase() + session.slice(1)}`;
    document.getElementById('session-description').textContent = `${results.length} drivers classified`;

    const winner = results.find(driver => driver.position === 1);
    document.getElementById('race-winner').textContent = winner?.driver_name ?? '--';
    document.getElementById('race-winner-card').classList.toggle('hidden', !winner);

    const tbody = document.getElementById('race-table-body');
    tbody.innerHTML = results.map(driver => {
        const finished = String(driver.status).toLowerCase() === 'finished';
        return `
            <tr class="hover:bg-[#232734]/50 transition-colors border-b border-[#232734]/50 group">
                <td class="py-3 px-3 md:px-4 font-black font-header text-2xl italic ${driver.position <= 3 ? 'text-[#0090D0]' : 'text-slate-600'}">
                    ${driver.position}
                </td>
                <td class="py-3 px-3 md:px-4 font-header font-bold text-slate-500 text-lg">
                    ${driver.race_number ?? '-'}
                </td>
                <td class="py-3 px-3 md:px-4">
                    <div class="font-bold text-white uppercase tracking-wider text-sm">${driver.driver_name}</div>
                </td>
                <td class="py-3 px-3 md:px-4 text-slate-400 text-xs font-bold uppercase">
                    ${driver.team_name ?? '-'}
                </td>
                <td class="py-3 px-3 md:px-4 text-center font-header font-bold text-slate-300 text-lg">
                    ${driver.num_laps ?? '-'}
                </td>
                <td class="py-3 px-3 md:px-4 font-header font-bold text-slate-400 text-base">
                    ${driver.best_lap_time ?? '-'}
                </td>
                <td class="py-3 px-3 md:px-4 font-header font-bold text-slate-400 text-base">
                    ${driver.total_race_time ?? '-'}
                </td>
                <td class="py-3 px-3 md:px-4 text-center">
                    <span class="inline-flex px-3 py-1 slanted text-[10px] font-black uppercase tracking-widest ${finished ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}">
                        ${driver.status ?? '-'}
                    </span>
                </td>
                <td class="py-3 px-3 md:px-4 text-right font-black font-header text-2xl italic group-hover:text-[#0090D0] transition-colors ${driver.points > 0 ? 'text-white' : 'text-slate-600'}">
                    ${driver.points ?? 0}
                </td>
            </tr>
        `;
    }).join('');
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