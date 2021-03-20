export default "@lib/ref";

const fn = () => 42;
export const X = () => fn();
export const Y = () => fn() + 1;
