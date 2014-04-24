'use strict';

angular.module('tnpldraftApp')
  .directive('rosterList', ['draftState', function (draftState) {
    return {
      templateUrl: 'scripts/directives/rosterlist.html',
      restrict: 'E',
			scope: {},
      link: function postLink(scope, element, attrs) {
				scope.team = undefined;

				scope.$watch(function() {
					return draftState.teams;
				}, function(teams) {
					scope.teams = teams;
				});

				scope.$watch(function() {
					return draftState.currentTeam;
				}, function() {
					if (scope.team === undefined) {
						scope.team = draftState.currentTeam;
					}
					scope.currentTeam = draftState.currentTeam;
				});

				scope.onPosClick = function(pos) {
					if (scope.selectedPosition === pos) {
						scope.selectedPosition = 'default';
					} else {
						scope.selectedPosition = pos;
					}
				}

				scope.$watch('team', function(a,b) {
					scope.selectedPosition = 'default';
				});
      }
    };
  }]);
