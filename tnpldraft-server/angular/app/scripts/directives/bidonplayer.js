'use strict';

angular.module('tnpldraftApp')
  .directive('bidOnPlayer', ['$interval', 'draftSocket', 'draftState', function ($interval, draftSocket, draftState) {
    return {
      templateUrl: 'scripts/directives/bidonplayer.html',
      restrict: 'E',
			scope: {
				auction: '=',
				onSubmit: '&',
			},
      link: function postLink(scope, element, attrs) {
				scope.$watch(function() {
					return draftState.currentTeam;
				}, function() {
					scope.team = draftState.currentTeam;
				});

				scope.onClick = function() {
					scope.onSubmit({player: scope.auction.player, bid: scope.selectedBid});
				}

				scope.$watch('auction.bid', function() {
					scope.auction.canBid = true;
					scope.auction.error = {
						hasRoomFor: !draftState.currentTeam.hasRoomFor(scope.auction.player),
						maxBid: scope.auction.bid >= draftState.currentTeam.maxBid(),
					};
					for (var errorKey in scope.auction.error) {
						if (scope.auction.error[errorKey]) {
							scope.auction.canBid = false;
						}
					}
					console.log("canBid: " + scope.auction.canBid);
					console.log("hasRoomFor: " + scope.auction.error.hasRoomFor);
					console.log("maxBid: " + scope.auction.error.maxBid);
					if (scope.auction && scope.auction.bid) {
						scope.minBid = scope.auction.bid + 50;
					} else {
						scope.minBid = 50;
					}
				});

				var updateSecondsLeftInterval;
				scope.$watch('auction.end_time', function() {
					if (updateSecondsLeftInterval) {
						$interval.cancel(updateSecondsLeftInterval);
					}
					if (scope.auction === undefined) {
						return;
					}
					scope.serverEndTime = +new Date(scope.auction.end_time);
					var serverTime = +draftSocket.serverTime();
					if (scope.serverEndTime > serverTime) {
						updateSecondsLeftInterval = $interval(updateSecondsLeft, 200);
					}
				});

				function updateSecondsLeft() {
					scope.secsLeft = Math.ceil((scope.serverEndTime - draftSocket.serverTime()) / 1000).toFixed(0);
				}
      }
    };
  }]);
