let raceIndex = [];
let loadedRaceDetails = {}; // Cache geladen race-details in geheugen
let chartInstance = null;

document.addEventListener("DOMContentLoaded", () => {
  initEventListeners();
  fetchRaceIndex();
});

function initEventListeners() {
  document.getElementById('nav-overview').addEventListener('click', showOverview);
  
  document.getElementById('mobile-menu-btn').addEventListener('click', () => {
    document.getElementById('sidebar').classList.toggle('-translate-x-full');
  });
}

function fetchRaceIndex() {
  fetch('./data/races.json')
    .then(res => {
      if (!res.ok) throw new Error("Index races.json niet gevonden");
      return res.json();
    })
    .then(data => {
      raceIndex = data;
      initDashboard();
    })
    .catch(err => {
      console.warn("Laden van data/races.json mislukt. Geen data beschikbaar.", err);
      initDashboard();
    });
}

function initDashboard() {
  setupSidebar();
  showOverview();
}

function setupSidebar() {
  const raceList = document.getElementById('race-list');
  raceList.innerHTML = '';

  raceIndex.forEach((raceSummary) => {
    const uid = raceSummary.SessionUID;
    const trackTitle = raceSummary.TrackName || `Race ${uid}`;
    
    const btn = document.createElement('button');
    btn.id = `nav-race-${uid}`;
    btn.className = 'w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-slate-400 hover:text-white hover:bg-f1-card text-sm transition-all group';
    btn.onclick = () => loadAndShowRace(uid);
    btn.innerHTML = `
      <div class="flex items-center space-x-2">
        <span class="text-xs">🏁</span>
        <span class="font-medium text-xs md:text-sm">${trackTitle}</span>
      </div>
      <span class="text-[10px] text-slate-500 group-hover:text-slate-300 font-mono">${raceSummary.NumCars || 0} Auto's</span>
    `;
    raceList.appendChild(btn);
  });
}

async function showOverview() {
  document.getElementById('view-overview').classList.remove('hidden');
  document.getElementById('view-race').classList.add('hidden');
  updateActiveNav('nav-overview');

  await ensureAllRaceDetailsLoaded();

  const standings = calculateStandings();
  renderStatCards(standings);
  renderStandingsTable(standings);
  renderChart(standings);
}

async function loadAndShowRace(sessionUID) {
  let raceDetails = loadedRaceDetails[sessionUID];

  if (!raceDetails) {
    try {
      const res = await fetch(`./data/race_${sessionUID}.json`);
      if (!res.ok) throw new Error(`Race ${sessionUID} niet gevonden`);
      raceDetails = await res.json();
      loadedRaceDetails[sessionUID] = raceDetails; // Sla op in cache
    } catch (err) {
      console.error(`Fout bij het laden van race_${sessionUID}.json`, err);
      return;
    }
  }

  renderRaceView(raceDetails, sessionUID);
}

function renderRaceView(race, sessionUID) {
  document.getElementById('view-overview').classList.add('hidden');
  document.getElementById('view-race').classList.remove('hidden');
  updateActiveNav(`nav-race-${sessionUID}`);

  document.getElementById('sidebar').classList.add('-translate-x-full');

  document.getElementById('race-title').innerText = race.TrackName || `Race ${sessionUID}`;
  
  const dateVal = race.RecordedAt;
  const dateStr = dateVal ? new Date(dateVal).toLocaleDateString() : '';
  document.getElementById('race-date').innerText = dateStr ? `Datum: ${dateStr}` : '';
  
  const drivers = race.Drivers || [];
  const winner = drivers.find(r => r.Position === 1);
  document.getElementById('race-winner').innerText = winner ? winner.Name : 'Onbekend';

  const tbody = document.getElementById('race-table-body');
  tbody.innerHTML = drivers.map(r => `
    <tr class="hover:bg-f1-card/50 transition-colors">
      <td class="py-3 px-3 md:px-4 font-black ${r.Position <= 3 ? 'text-amber-400' : 'text-slate-400'}">${r.Position}</td>
      <td class="py-3 px-3 md:px-4 font-mono text-slate-500">${r.RaceNumber}</td>
      <td class="py-3 px-3 md:px-4 font-bold text-white">${r.Name}</td>
      <td class="py-3 px-3 md:px-4 text-center font-mono">${r.NumLaps}</td>
      <td class="py-3 px-3 md:px-4 font-mono text-xs">${r.TotalRaceTime}</td>
      <td class="py-3 px-3 md:px-4 font-mono text-xs text-purple-400">${r.BestLapTime}</td>
      <td class="py-3 px-3 md:px-4 text-center">
        <span class="px-2 py-0.5 rounded-full text-[10px] font-bold ${r.ResultStatusText === 'Finished' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'}">
          ${r.ResultStatusText === 'Finished' ? 'Finished' : 'DNF'}
        </span>
      </td>
      <td class="py-3 px-3 md:px-4 text-right font-black text-f1-red">+${r.Points}</td>
    </tr>
  `).join('');
}

async function ensureAllRaceDetailsLoaded() {
  const fetchPromises = raceIndex.map(async (summary) => {
    const uid = summary.SessionUID;
    if (!loadedRaceDetails[uid]) {
      try {
        const res = await fetch(`./data/race_${uid}.json`);
        if (res.ok) {
          loadedRaceDetails[uid] = await res.json();
        }
      } catch (err) {
        console.warn(`Kon details niet laden voor UID: ${uid}`);
      }
    }
  });

  await Promise.all(fetchPromises);
}

function calculateStandings() {
  const driverMap = {};
  
  Object.values(loadedRaceDetails).forEach(race => {
    const drivers = race.Drivers || [];
    drivers.forEach(d => {
      if (!driverMap[d.Name]) {
        driverMap[d.Name] = { name: d.Name, points: 0, wins: 0, podiums: 0, dnfs: 0 };
      }
      driverMap[d.Name].points += d.Points;
      if (d.Position === 1) driverMap[d.Name].wins++;
      if (d.Position <= 3) driverMap[d.Name].podiums++;
      if (d.ResultStatusText !== 'Finished') driverMap[d.Name].dnfs++;
    });
  });

  return Object.values(driverMap).sort((a, b) => b.points - a.points);
}

function renderStatCards(standings) {
  const leader = standings[0] || { name: '--', points: 0, wins: 0 };
  const totalRaces = raceIndex.length;

  const container = document.getElementById('stats-cards');
  container.innerHTML = `
    <div class="bg-f1-card p-3 md:p-4 rounded-xl border border-f1-border">
      <div class="text-[9px] md:text-[10px] font-bold text-slate-400 uppercase tracking-wider">Koploper</div>
      <div class="text-sm md:text-lg font-black text-white mt-1 truncate">${leader.name}</div>
      <div class="text-xs text-f1-red font-bold mt-0.5">${leader.points} Pts</div>
    </div>
    <div class="bg-f1-card p-3 md:p-4 rounded-xl border border-f1-border">
      <div class="text-[9px] md:text-[10px] font-bold text-slate-400 uppercase tracking-wider">Races Verwerkt</div>
      <div class="text-sm md:text-lg font-black text-white mt-1">${totalRaces}</div>
      <div class="text-xs text-emerald-400 font-bold mt-0.5">Voltooide sessies</div>
    </div>
    <div class="bg-f1-card p-3 md:p-4 rounded-xl border border-f1-border">
      <div class="text-[9px] md:text-[10px] font-bold text-slate-400 uppercase tracking-wider">Meeste Overwinningen</div>
      <div class="text-sm md:text-lg font-black text-white mt-1 truncate">${leader.name}</div>
      <div class="text-xs text-amber-400 font-bold mt-0.5">${leader.wins} Wins</div>
    </div>
    <div class="bg-f1-card p-3 md:p-4 rounded-xl border border-f1-border">
      <div class="text-[9px] md:text-[10px] font-bold text-slate-400 uppercase tracking-wider">Actieve Coureurs</div>
      <div class="text-sm md:text-lg font-black text-white mt-1">${standings.length}</div>
      <div class="text-xs text-slate-400 font-bold mt-0.5">In klassement</div>
    </div>
  `;
}

function renderStandingsTable(standings) {
  const tbody = document.getElementById('standings-table-body');
  tbody.innerHTML = standings.map((d, i) => `
    <tr class="hover:bg-f1-card/50 transition-colors">
      <td class="py-3 px-3 md:px-4 font-black ${i === 0 ? 'text-amber-400' : 'text-slate-400'}">P${i + 1}</td>
      <td class="py-3 px-3 md:px-4 font-bold text-white">${d.name}</td>
      <td class="py-3 px-3 md:px-4 text-center font-mono">${d.wins}</td>
      <td class="py-3 px-3 md:px-4 text-center font-mono">${d.podiums}</td>
      <td class="py-3 px-3 md:px-4 text-center font-mono text-rose-400">${d.dnfs}</td>
      <td class="py-3 px-3 md:px-4 text-right font-black text-white text-sm md:text-base">${d.points}</td>
    </tr>
  `).join('');
}

function renderChart(standings) {
  const ctx = document.getElementById('pointsChart').getContext('2d');
  if (chartInstance) chartInstance.destroy();

  const topDrivers = standings.slice(0, 4);
  const labels = raceIndex.map((r) => r.TrackName || `Race ${r.SessionUID}`);

  const datasets = topDrivers.map((driver, idx) => {
    const colors = ['#e10600', '#3b82f6', '#10b981', '#f59e0b'];
    let accumulatedPoints = 0;
    
    const pointsData = raceIndex.map(summary => {
      const raceDetail = loadedRaceDetails[summary.SessionUID];
      if (raceDetail && raceDetail.Drivers) {
        const res = raceDetail.Drivers.find(r => r.Name === driver.name);
        accumulatedPoints += res ? res.Points : 0;
      }
      return accumulatedPoints;
    });

    return {
      label: driver.name,
      data: pointsData,
      borderColor: colors[idx % colors.length],
      backgroundColor: colors[idx % colors.length],
      borderWidth: 2,
      tension: 0.3
    };
  });

  chartInstance = new Chart(ctx, {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { labels: { color: '#94a3b8', font: { size: 11 } } }
      },
      scales: {
        x: { grid: { color: '#2d2d38' }, ticks: { color: '#94a3b8', font: { size: 10 } } },
        y: { grid: { color: '#2d2d38' }, ticks: { color: '#94a3b8', font: { size: 10 } } }
      }
    }
  });
}

function updateActiveNav(activeId) {
  const allBtns = document.querySelectorAll('#sidebar button');
  allBtns.forEach(btn => {
    if (btn.id === activeId) {
      btn.classList.add('bg-f1-red', 'text-white', 'shadow-lg', 'shadow-f1-red/20');
      btn.classList.remove('text-slate-400', 'hover:bg-f1-card');
    } else {
      btn.classList.remove('bg-f1-red', 'text-white', 'shadow-lg', 'shadow-f1-red/20');
      btn.classList.add('text-slate-400', 'hover:bg-f1-card');
    }
  });
}