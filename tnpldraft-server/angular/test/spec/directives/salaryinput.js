'use strict';

describe('Directive: salaryInput', function () {

  // load the directive's module
  beforeEach(module('tnpldraftApp'));

  var element,
    scope;

  beforeEach(inject(function ($rootScope) {
    scope = $rootScope.$new();
  }));

  it('should make hidden element visible', inject(function ($compile) {
    element = angular.element('<salary-input></salary-input>');
    element = $compile(element)(scope);
    expect(element.text()).toBe('this is the salaryInput directive');
  }));
});
