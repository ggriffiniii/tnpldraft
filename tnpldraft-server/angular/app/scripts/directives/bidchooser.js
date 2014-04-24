'use strict';

angular.module('tnpldraftApp')
  .directive('bidChooser', function ($interval) {
    return {
      templateUrl: 'scripts/directives/bidchooser.html',
      restrict: 'E',
			scope: {
				minBid: '=',
				maxBid: '=',
				selected: '='
			},
      link: function postLink(scope, element, attrs) {
				scope.$watch('maxBid', function() {
					scope.allBids = [];
					for (var i = 0; i <= scope.maxBid; i += 50) {
						scope.allBids.push(i);
					}
				});
				scope.$watch('minBid', function() {
					if (scope.minBid === undefined) {
						scope.minBid = 50;
					}
					if (scope.selected < scope.minBid) {
						scope.selected = -1;  // negative values inidicate no selection.
					}
					var i = scope.minBid / 50;
					var buttonWidth = 71;
					scope.offset = -buttonWidth * i - 2;
				});
				scope.select = function(bid) {
					scope.selected = bid;
				}
      }
    };
  });
