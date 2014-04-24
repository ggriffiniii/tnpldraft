'use strict';

angular.module('tnpldraftApp')
  .directive('salaryInput', ['draftState', function (draftState) {
    return {
      template: '<input type="text" class="form-control" ng-model="inputStr"></input>',
      restrict: 'E',
			require: 'ngModel',
      link: function postLink(scope, element, attrs, ngModelCtrl) {
				scope.inputStr = '';
				ngModelCtrl.$formatters.push(function(modelVal) {
					return modelVal/100.00;
				});
				ngModelCtrl.$render = function() {
					if (ngModelCtrl.$viewValue < 0) {
						scope.inputStr = '$';
					} else {
						scope.inputStr = '$' + ngModelCtrl.$viewValue.toFixed(2);
					}
				}
				ngModelCtrl.$parsers.push(function(viewVal) {
					if (isNaN(viewVal)) {
						ngModelCtrl.$setValidity('required', false);
						ngModelCtrl.$setValidity('nan', false);
						ngModelCtrl.$setValidity('multiple', true);
						ngModelCtrl.$setValidity('tooLow', true);
						ngModelCtrl.$setValidity('aboveMax', true);
						return -100;
					} else if (viewVal < 0) {
						ngModelCtrl.$setValidity('required', false);
						ngModelCtrl.$setValidity('nan', true);
						ngModelCtrl.$setValidity('multiple', true);
						ngModelCtrl.$setValidity('tooLow', true);
						ngModelCtrl.$setValidity('aboveMax', true);
						return -100;
					} else {
						ngModelCtrl.$setValidity('required', true);
						ngModelCtrl.$setValidity('nan', true);
						var x = Math.floor(viewVal * 100);
						if (x % 50 != 0) {
							ngModelCtrl.$setValidity('multiple', false);
						} else {
							ngModelCtrl.$setValidity('multiple', true);
						}
						if (x < 50) {
							ngModelCtrl.$setValidity('tooLow', false);
						} else {
							ngModelCtrl.$setValidity('tooLow', true);
						}
						if (x > draftState.currentTeam.maxBid()) {
							ngModelCtrl.$setValidity('aboveMax', false);
						} else {
							ngModelCtrl.$setValidity('aboveMax', true);
						}
						return x;
					}
				});
				
        scope.$watch('inputStr', function() {
					var x = scope.inputStr.trim();
					if (x.substr(0,1) === '$') {
						x = x.substr(1);
					}
					ngModelCtrl.$setViewValue(parseFloat(x));
				});
      }
    };
  }]);
