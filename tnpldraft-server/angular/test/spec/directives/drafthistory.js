'use strict';

describe('Directive: draftHistory', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));

  var element,
    scope,
		draft;

  beforeEach(inject(function ($compile, $rootScope, draftState) {
    scope = $rootScope.$new();
		element = angular.element("<draft-history></draft-history>");
		$compile(element)(scope);
		draft = draftState;
  }));

	it('should start empty', function() {
		expect(element.find('table').length).toBe(1);
	});
});
