'use strict';

describe('Directive: rosterList', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));

  var element,
    scope;

  beforeEach(inject(function ($rootScope) {
    scope = $rootScope.$new();
  }));

  it('should make hidden element visible', inject(function ($compile) {
    element = angular.element('<roster-list></roster-list>');
    element = $compile(element)(scope);
    expect(element.text()).toBe('this is the rosterList directive');
  }));
});
