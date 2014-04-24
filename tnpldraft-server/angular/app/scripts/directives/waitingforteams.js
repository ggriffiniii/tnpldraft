'use strict';

angular.module('tnpldraftApp')
  .directive('waitingForTeams', ['draftState', function (draftState) {
    return {
      templateUrl: 'scripts/directives/waitingforteams.html',
			scope: {},
      restrict: 'E',
      link: function postLink(scope, element, attrs) {
				scope.$watch(function() {
					return draftState.teams;
				}, function() {
					scope.teams = draftState.teams;
				});
      }
    };
  }]);
