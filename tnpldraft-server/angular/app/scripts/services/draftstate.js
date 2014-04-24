'use strict';

angular.module('tnpldraftApp')
  .factory('draftState', function () {

		var ALL_POS = ['P', 'C', '1B', '2B', '3B', 'SS', 'MI', 'CI', 'OF', 'U'];

		function newDraft() {
			var draft = {};
			draft.initialized = false;
			draft.id = undefined;
			draft.name = undefined;
			draft.teams = [];
			draft.teamsById = {};
			draft.picks = [];
			draft.requiredPositions = {};
			draft.playerCount = 0;
			draft.salaryCap = undefined;
			draft.currentTeam = undefined;
			draft.ownedPlayersById = {};

			draft.initDraft = function (draftInfo) {
				//draft.id = draftInfo.id;
				draft.name = draftInfo.name;
				draft.salaryCap = draftInfo.salary_cap;
				draft.requiredPositions = angular.copy(draftInfo.positions);
				draftInfo.teams.forEach(function(teamInfo) {
					var team = draft.newTeam(teamInfo);
					draft.addTeam(team);
				});
				draftInfo.picks.forEach(function(pickInfo) {
					draft.addPick(pickInfo);
				});
				angular.forEach(draft.requiredPositions, function(val) {
					draft.playerCount += val;
				});
				draft.currentTeam = draft.getTeam(draftInfo.team);
				draft.initialized = true;
			}

			draft.addTeam = function(team) {
					draft.teams.push(team);
					draft.teamsById[team.id] = team;
					return team;
			}

			draft.getTeam = function(teamId) {
				return draft.teamsById[teamId];
			}

			draft.addPick = function(pickInfo) {
				var winningTeam = draft.getTeam(pickInfo.winning_team);
				draft.picks.push({
					winning_team: winningTeam,
					offering_team: draft.getTeam(pickInfo.offering_team),
					player: winningTeam.getPlayer(pickInfo.player.id)
				});
			}

			draft.newTeam = function(teamInfo) {
				var team = {};
				team.id = teamInfo.id;
				team.name = teamInfo.name;
				team.players = [];
				team.playersById = {};
				team.connected = false;
				team.draftablePositions = ALL_POS;
				var validRosters = {};
				getValidRosters();

				team.addPlayer = function(player) {
					team.players.push(player);
					team.playersById[player.id] = player;
					draft.ownedPlayersById[player.id] = team;
					getValidRosters();
					return player;
				}

				team.getPlayer = function(playerId) {
					return team.playersById[playerId];
				}

				team.numPlayers = function() {
					return team.players.length;
				}

				team.requiredNumPlayers = function() {
					return draft.playerCount;
				}

				// Sets validRosters. This is fairly compute intensive and only changes
				// when new players are added, so on every addPlayer we recompute and
				// cache the results in validRosters.
				function getValidRosters() {
					var tempRosters = {};
					function rosterHelper(assigned, remaining) {
						if (remaining.length == 0) {
							if (!('default' in tempRosters)) {
								tempRosters['default'] = angular.copy(assigned);
							}
							angular.forEach(assigned, function(players,pos) {
								if (players.length < draft.requiredPositions[pos]) {
									if (!(pos in tempRosters)) {
										tempRosters[pos] = angular.copy(assigned);
									}
								}
							});
							return;
						}
						var player = remaining[0];
						player.positions.forEach(function(pos) {
							if (assigned[pos].length < draft.requiredPositions[pos]) {
								// Room for player at pos.
								assigned[pos].push(player);
								rosterHelper(assigned, remaining.slice(1));
								assigned[pos].pop();
							}
						});
					}
					var assigned = {};
					angular.forEach(draft.requiredPositions, function(unused,pos) {
						assigned[pos] = [];
					});
					// rosterHelper populates tempRosters.
					rosterHelper(assigned, team.players);
					validRosters = {};
					angular.forEach(tempRosters, function(roster,pos) {
						var newRoster = [];
						angular.forEach(ALL_POS, function(position) {
							var count = draft.requiredPositions[position];
							roster[position].forEach(function(player) {
								newRoster.push({pos: position, player: player});
								--count;
							});
							for (;count > 0; --count) {
								newRoster.push({pos: position, player: {}});
							}
						});
						validRosters[pos] = newRoster;
					});
					team.draftablePositions = ALL_POS.filter(function(pos) {
						return pos in validRosters;
					});
				}

				team.getRoster = function(availablePos) {
					return validRosters[availablePos || 'default'];
				}

				team.totalSalary = function() {
					return team.players.reduce(function(prev, curr) {
						return prev + curr.salary;
					}, 0);
				}

				team.remainingSalary = function() {
					return draft.salaryCap - team.totalSalary();
				}

				team.playersNeeded = function() {
					return draft.playerCount - team.players.length;
				}

				team.maxBid = function() {
					return 50 + team.remainingSalary() - team.playersNeeded() * 50;
				}

				team.hasRoomFor = function(player) {
					for (var i = 0; i < player.positions.length; ++i) {
						for (var j = 0; j < team.draftablePositions.length; ++j) {
							if (player.positions[i] === team.draftablePositions[j]) {
								return true;
							}
						}
					}
					return false;
				}

				// Initialize the players on the team.
				teamInfo.players.forEach(function(playerInfo) {
					var player = draft.newPlayer(playerInfo);
					team.addPlayer(player);
				});

				return team;
			}

			draft.newPlayer = function(playerInfo) {
				var player = {};
				player.id = playerInfo.id;
				player.firstname = playerInfo.firstname;
				player.lastname = playerInfo.lastname;
				player.fullname = playerInfo.firstname + ' ' + playerInfo.lastname;
				player.mlbteam = playerInfo.mlbteam;
				player.positions = playerInfo.positions;
				player.salary = playerInfo.salary;

				player.team = function() {
					if (player.id in draft.ownedPlayersById) {
						return draft.ownedPlayersById[player.id];
					}
					return undefined;
				}

				return player;
			}

			return draft;
		}

		return newDraft();
  });
