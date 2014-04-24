'use strict';

describe('Directive: waitingForTeams', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));

  var element,
    scope;

  beforeEach(inject(function ($rootScope) {
    scope = $rootScope.$new();
  }));

  it('should make hidden element visible', inject(function ($compile) {
    element = angular.element('<waiting-for-teams></waiting-for-teams>');
    element = $compile(element)(scope);
    expect(element.text()).toBe('this is the waitingForTeams directive');
  }));
});
