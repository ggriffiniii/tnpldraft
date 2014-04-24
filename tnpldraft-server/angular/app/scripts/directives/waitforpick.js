'use strict';

angular.module('tnpldraftApp')
  .directive('waitForPick', function () {
    return {
      templateUrl: 'scripts/directives/waitforpick.html',
			scope: {
				team: '=',
			},
      restrict: 'E',
    };
  });
