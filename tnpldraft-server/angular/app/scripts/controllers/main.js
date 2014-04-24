'use strict';

angular.module('tnpldraftApp')
  .controller('MainCtrl', ['$scope', 'draftSocket', 'draftState', function ($scope, draftSocket, draftState) {
		$scope.state = 'INIT';

		$scope.$watchCollection(function() {
			return draftState.teams;
		}, function(teams) {
			$scope.teams = teams;
		});

    draftSocket.onmessage = function(e) {
			var msg = angular.fromJson(e.data);
			if (msg.type === 'DraftSummary') {
				draftState.initDraft(msg.data);
				$scope.state = 'CONNECTED';
			} else if (msg.type === 'TeamJoinLeaveMessage') {
				msg.data.connected.forEach(function(id) {
					draftState.getTeam(id).connected = true;
				});
				msg.data.disconnected.forEach(function(id) {
					draftState.getTeam(id).connected = false;
				});
			} else if (msg.type === 'WaitingForPick') {
				if (draftState.getTeam(msg.data.team) === draftState.currentTeam) {
					$scope.state = 'OFFER_PICK';
				} else {
					$scope.state = 'WAITING_FOR_PICK';
					$scope.pickingTeam = draftState.getTeam(msg.data.team);
				}
			} else if (msg.type === 'PlayerRejected') {
				if ($scope.state !== 'OFFER_PICK') {
					return;
				}
				console.log("setting player_errror");
				$scope.player_error = msg.data;		
			} else if (msg.type === 'Auction') {
				$scope.state = 'AUCTION_IN_PROGRESS';
				$scope.auction = msg.data;
				$scope.auction.player = draftState.newPlayer(msg.data.player);
				$scope.auction.team = draftState.getTeam(msg.data.team);
			} else if (msg.type === 'AuctionComplete') {
				var team = draftState.getTeam(msg.data.winning_team);
				team.addPlayer(draftState.newPlayer(msg.data.player));
				draftState.addPick(msg.data);
			} else {
				console.log("don't understand message");
			}
    }

		$scope.pickPlayer = function(player, bid) {
			draftSocket.send(angular.toJson({
				type: 'Pick',
				data: {
					player: player,
					bid: bid
				}
			}));
		};

		$scope.bidOnPlayer = function(player, bid) {
			draftSocket.send(angular.toJson({
				type: 'Bid',
				data: {
					player: player,
					bid: bid
				}
			}));
		};

		$scope.stateSummary = function() {
			if ($scope.state === 'INIT') {
				return 'Connecting to draft';
			} else if ($scope.state === 'CONNECTED') {
				return 'Waiting for all teams to join the draft';
			} else if ($scope.state === 'OFFER_PICK') {
				return 'You need to choose a player';
			} else if ($scope.state === 'WAITING_FOR_PICK') {
				return 'Waiting for next player';
			} else if ($scope.state === 'AUCTION_IN_PROGRESS') {
				return 'Auction in progress';
			}
			return 'Unknown state';
		}
  }]);
