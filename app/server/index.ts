import express from "express";
import cors from "cors";
import appointmentsRouter from "./routes/appointments";
import patientsRouter from "./routes/patients";
import webhookRouter from "./routes/webhook";

const app = express();
const PORT = 3001;

app.use(cors());
app.use(express.json());

app.get("/api/health", (_req, res) => res.json({ status: "ok" }));
app.use("/api/appointments", appointmentsRouter);
app.use("/api/patients", patientsRouter);
app.use("/api/webhook", webhookRouter);

app.listen(PORT, () => {
  console.log(`🦷 Dental AI Receptionist API running on http://localhost:${PORT}`);
});
