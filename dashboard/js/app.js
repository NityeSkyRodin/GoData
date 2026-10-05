let raceIndex = [];
let championship = [];
let currentRace = null;
let currentSession = 'race';


document.addEventListener('DOMContentLoaded', async () => {
    initEventListeners();

    try {
        await loadChampionship();
        await loadRaceIndex();

        renderChampionship();
        setupSidebar();

    } catch (error) {
        console.error('Failed to initialize dashboard:', error);
    }
});


function initEventListeners() {
    document
        .getElementById('nav-overview')
        .addEventListener('click', showOverview);

    document
        .getElementById('mobile-menu-btn')
        .addEventListener('click', () => {
            document
                .getElementById('sidebar')
                .classList.toggle('-translate-x-full');
        });
}


async function loadChampionship() {
    const response = await fetch('./data/championship.json');

    if (!response.ok) {
        throw new Error('championship.json not found');
    }

    championship = await response.json();
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

        button.className =
            'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg ' +
            'text-slate-400 hover:text-white hover:bg-f1-card ' +
            'text-sm transition-all group';

        button.innerHTML = `
            <span class="text-xs">🏁</span>
            <span class="font-medium">${race.track_name}</span>
        `;

        button.addEventListener('click', () => {

            loadRace(race);
        });

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
        <div class="bg-f1-card p-4 rounded-xl border border-f1-border">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                Championship Leader
            </div>

            <div class="text-lg font-black text-white mt-1 truncate">
                ${leader?.driver_name ?? '--'}
            </div>

            <div class="text-xs text-f1-red font-bold mt-1">
                ${leader?.points ?? 0} points
            </div>
        </div>


        <div class="bg-f1-card p-4 rounded-xl border border-f1-border">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                Races Completed
            </div>

            <div class="text-lg font-black text-white mt-1">
                ${raceIndex.length}
            </div>

            <div class="text-xs text-emerald-400 font-bold mt-1">
                Race weekends
            </div>
        </div>


        <div class="bg-f1-card p-4 rounded-xl border border-f1-border">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                Drivers
            </div>

            <div class="text-lg font-black text-white mt-1">
                ${championship.length}
            </div>

            <div class="text-xs text-slate-400 font-bold mt-1">
                Championship
            </div>
        </div>


        <div class="bg-f1-card p-4 rounded-xl border border-f1-border">
            <div class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                Points Leader
            </div>

            <div class="text-lg font-black text-white mt-1">
                ${leader?.points ?? 0}
            </div>

            <div class="text-xs text-slate-400 font-bold mt-1">
                Total points
            </div>
        </div>
    `;

    document.getElementById('season-status').textContent =
        `${raceIndex.length} race${raceIndex.length === 1 ? '' : 's'} completed`;
}


function renderStandings() {
    const tbody = document.getElementById('standings-table-body');

    tbody.innerHTML = championship.map((driver, index) => `
        <tr class="hover:bg-f1-card/50 transition-colors">

            <td class="py-3 px-4 font-black ${
                index === 0
                    ? 'text-amber-400'
                    : 'text-slate-500'
            }">
                ${index + 1}
            </td>

            <td class="py-3 px-4 font-bold text-white">
                ${driver.driver_name}
            </td>

            <td class="py-3 px-4 text-slate-400">
                ${driver.team_name}
            </td>

            <td class="py-3 px-4 text-right font-black text-white">
                ${driver.points}
            </td>

        </tr>
    `).join('');
}


async function loadRace(race) {
    try {
        const response = await fetch(`./data/races/${race.file}`);

        if (!response.ok) {
            throw new Error(`Could not load ${race.file}`);
        }

        currentRace = await response.json();

        renderRace(currentRace, race);

    } catch (error) {
        console.error('Failed to load race:', error);
    }
}


function renderRace(race, raceIndexEntry) {
    document.getElementById('view-overview').classList.add('hidden');
    document.getElementById('view-race').classList.remove('hidden');

    document
        .getElementById('sidebar')
        .classList.add('-translate-x-full');

    updateActiveNav(null);

    document.getElementById('race-title').textContent =
        race.track_name;

    document.getElementById('race-subtitle').textContent =
        `Track ID ${race.track_id}`;

    renderSessionTabs(race);

    selectSession(getFirstAvailableSession(race));
}


function renderSessionTabs(race) {
    const container = document.getElementById('session-tabs');

    const sessions = [
        {
            key: 'qualifying',
            label: 'Qualifying'
        },
        {
            key: 'sprint',
            label: 'Sprint'
        },
        {
            key: 'race',
            label: 'Race'
        }
    ];

    container.innerHTML = sessions
        .filter(session => race[session.key]?.length)
        .map(session => `
            <button
                data-session="${session.key}"
                class="
                    session-tab
                    px-4 py-2 rounded-lg
                    text-xs font-bold uppercase tracking-wider
                    bg-f1-card border border-f1-border
                    text-slate-400
                    hover:text-white hover:border-slate-600
                    transition-all
                "
            >
                ${session.label}
            </button>
        `)
        .join('');

    container
        .querySelectorAll('.session-tab')
        .forEach(button => {
            button.addEventListener('click', () => {
                selectSession(button.dataset.session);
            });
        });
}


function getFirstAvailableSession(race) {
    if (race.race?.length) {
        return 'race';
    }

    if (race.sprint?.length) {
        return 'sprint';
    }

    return 'qualifying';
}


function selectSession(session) {
    currentSession = session;

    const results = currentRace?.[session] ?? [];

    document
        .querySelectorAll('.session-tab')
        .forEach(button => {
            const active = button.dataset.session === session;

            button.classList.toggle('bg-f1-red', active);
            button.classList.toggle('text-white', active);

            button.classList.toggle('text-slate-400', !active);
        });

    renderSession(results, session);
}


function renderSession(results, session) {
    const title = document.getElementById('session-title');

    title.textContent =
        `${session.charAt(0).toUpperCase() + session.slice(1)} Classification`;

    document.getElementById('session-description').textContent =
        `${results.length} drivers classified`;

    const winner = results.find(driver => driver.position === 1);

    document.getElementById('race-winner').textContent =
        winner?.driver_name ?? '--';

    document
        .getElementById('race-winner-card')
        .classList.toggle('hidden', !winner);

    const tbody = document.getElementById('race-table-body');

    tbody.innerHTML = results.map(driver => {
        const finished = driver.status === 'finished';

        return `
            <tr class="hover:bg-f1-card/50 transition-colors">

                <td class="py-3 px-3 md:px-4 font-black ${
                    driver.position <= 3
                        ? 'text-amber-400'
                        : 'text-slate-400'
                }">
                    ${driver.position}
                </td>

                <td class="py-3 px-3 md:px-4 font-mono text-slate-500">
                    ${driver.race_number}
                </td>

                <td class="py-3 px-3 md:px-4">
                    <div class="font-bold text-white">
                        ${driver.driver_name}
                    </div>
                </td>

                <td class="py-3 px-3 md:px-4 text-slate-400">
                    ${driver.team_name}
                </td>

                <td class="py-3 px-3 md:px-4 text-center font-mono">
                    ${driver.num_laps}
                </td>

                <td class="py-3 px-3 md:px-4 font-mono text-xs">
                    ${driver.best_lap_time}
                </td>

                <td class="py-3 px-3 md:px-4 font-mono text-xs">
                    ${driver.total_race_time}
                </td>

                <td class="py-3 px-3 md:px-4 text-center">

                    <span class="
                        inline-flex px-2 py-0.5 rounded-full
                        text-[10px] font-bold
                        ${
                            finished
                                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                        }
                    ">
                        ${finished ? 'Finished' : 'DNF'}
                    </span>

                </td>

                <td class="py-3 px-3 md:px-4 text-right font-black ${
                    driver.points > 0
                        ? 'text-f1-red'
                        : 'text-slate-500'
                }">
                    ${driver.points}
                </td>

            </tr>
        `;
    }).join('');
}


function updateActiveNav(activeId) {
    document
        .querySelectorAll('#sidebar button')
        .forEach(button => {
            const active = button.id === activeId;

            button.classList.toggle('bg-f1-red', active);
            button.classList.toggle('text-white', active);
            button.classList.toggle('shadow-lg', active);
            button.classList.toggle('shadow-f1-red/20', active);

            if (!active) {
                button.classList.add('text-slate-400');
                button.classList.remove('text-white');
            }
        });
}