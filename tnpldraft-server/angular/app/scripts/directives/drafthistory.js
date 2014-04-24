'use strict';

angular.module('tnpldraftApp')
  .directive('draftHistory', ['draftState', function (draftState) {
    return {
      templateUrl: 'scripts/directives/drafthistory.html',
			scope: {},
      restrict: 'E',
      link: function postLink(scope, element, attrs) {
				scope.$watch(function() {
					return draftState.picks;
				}, function() {
					scope.picks = draftState.picks;
				});
      }
    };
  }]);
