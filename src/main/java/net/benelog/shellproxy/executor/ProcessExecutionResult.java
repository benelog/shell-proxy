package net.benelog.shellproxy.executor;

public class ProcessExecutionResult {
	private int exitCode;
	private String standardOutput;
	private String standardError;
	
	public int getExitCode() {
		return exitCode;
	}
	public void setExitCode(int exitCode) {
		this.exitCode = exitCode;
	}
	public String getStandardOutput() {
		return standardOutput;
	}
	public void setStandardOutput(String standardOutput) {
		this.standardOutput = standardOutput;
	}
	public String getStandardError() {
		return standardError;
	}
	public void setStandardError(String standardError) {
		this.standardError = standardError;
	}
	@Override
	public String toString() {
		return "[\n exitCode = " + exitCode + ", \n standardOutput = {\n" + standardOutput + "\n  }, \n standardError = {\n" + standardError + "\n  }\n]";
	}
}
