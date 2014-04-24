'use strict';

describe('Service: draftState', function () {

  // load the service's module
  beforeEach(module('tnpldraftApp'));

  // instantiate service
  var draftState;
  beforeEach(inject(function (_draftState_) {
    draftState = _draftState_;
  }));

  it('should do something', function () {
    expect(!!draftState).toBe(true);
  });

});
