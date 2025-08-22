import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("CounterModule", (m) => {
  const counter = m.contract("Counter");

  m.call(counter, "incBy", [5n]);
  m.call(counter, "incBy", [5n], {
    id: "second_call",
  });

  const x = m.staticCall(counter, "x", [], "x");
  console.log("x", x);

  return { counter };
});
