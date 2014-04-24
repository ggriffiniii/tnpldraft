'use strict';

angular.module('tnpldraftApp')
  .filter('dollars', function () {
    return function (input) {
			if (typeof(input) === 'number') {
				var num = input / 100.00;
				return '$' + num.toFixed(2);
			}
			return '';
    };
  });
