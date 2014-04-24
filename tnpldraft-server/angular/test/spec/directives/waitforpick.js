'use strict';

describe('Directive: waitForPick', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));

  var element,
    scope;

  beforeEach(inject(function ($rootScope) {
    scope = $rootScope.$new();
  }));

  it('should make hidden element visible', inject(function ($compile) {
    element = angular.element('<wait-for-pick></wait-for-pick>');
    element = $compile(element)(scope);
    expect(element.text()).toBe('this is the waitForPick directive');
  }));
});
