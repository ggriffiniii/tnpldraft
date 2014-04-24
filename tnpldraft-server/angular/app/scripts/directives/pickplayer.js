'use strict';

angular.module('tnpldraftApp')
  .directive('pickPlayer', ['$http', '$filter', 'draftState', function ($http, $filter, draftState) {
    return {
      templateUrl: 'scripts/directives/pickplayer.html',
      restrict: 'E',
			scope: {
				error: '=',
				onSubmit: '&',
			},
			link: function(scope) {
				scope.selection = {
					submitted: false,
					bid: 50,
					player: undefined,
					invalid: false,
					error: {
						hasRoomFor: false,
						freeAgent: false
					}
				};
				scope.$watch(function() {
					return draftState.currentTeam;
				}, function() {
					scope.team = draftState.currentTeam;
				});

				scope.getPlayers = function(search) {
					return $http.get('/api/draft/5/playerfilter', {
						params: {
							name: search
						}
					}).then(function(res) {
						return res.data.map(function(playerInfo) {
							return draftState.newPlayer(playerInfo);
						});
					});
				};

				scope.buttonText = function() {
					if (scope.selection.submitted) {
						return 'Submitted';
					}
					if (scope.selection.player === undefined) {
						return '';
					}
					return 'Offer ' + scope.selection.player.fullname + ' for ' + $filter('dollars')(scope.selection.bid);
				}

				scope.onClick = function() {
					scope.selection.submitted = true;
					scope.onSubmit({player: scope.selection.player, bid: scope.selection.bid});
				}

				scope.onSelect = function(player) {
					scope.selection = {
						submitted: false,
						bid: 50,
						player: player,
						invalid: false,
						error: {
							hasRoomFor: !scope.team.hasRoomFor(player),
							freeAgent: player.team() !== undefined,
						}
					};
					for (var errorKey in scope.selection.error) {
						if (scope.selection.error[errorKey]) {
							scope.selection.invalid = true;
							break;
						}
					}
				};
			}
    }
  }]);
