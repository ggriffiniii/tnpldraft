'use strict';

describe('Service: draftSocket', function () {

  // load the service's module
  beforeEach(module('tnpldraftApp'));

  // instantiate service
  var draftSocket;
  beforeEach(inject(function (_draftSocket_) {
    draftSocket = _draftSocket_;
  }));

  it('should do something', function () {
    expect(!!draftSocket).toBe(true);
  });

});
