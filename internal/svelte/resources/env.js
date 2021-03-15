self = this;
if (!self.performance) {
  self.performance = { now: () => 0 };
}

window = self;
