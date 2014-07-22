package net.benelog.shellproxy.executor;

import static org.hamcrest.CoreMatchers.*;
import static org.junit.Assert.*;

import java.io.OutputStream;
import java.io.StringWriter;

import net.benelog.shellproxy.executor.ShellExecutor;

import org.eclipse.jetty.io.WriterOutputStream;
import org.junit.Test;

/**
 * @author benelog
 */
public class ShellExecutorTest {
	int timeoutMillsecs = 6000;

	@Test
	public void exitCodeShouldBe0() {
		ShellExecutor cutycapt =  new ShellExecutor();
		
		int exitCode = cutycapt.execute("echo hello");
		
		assertThat(exitCode, is(0));
	}

	@Test
	public void exitCodeShouldNotBe0() {
		ShellExecutor shell =  new ShellExecutor(timeoutMillsecs);
		
		int exitCode = shell.execute("javac");
		assertThat(exitCode, is(not(0)));
	}
	
	@Test
	public void stdoutShouldBeCaptured() {
		
		StringWriter stdout = new StringWriter();
		StringWriter stderr = new StringWriter();
		ShellExecutor shell =  new ShellExecutor(timeoutMillsecs, stream(stdout), stream(stderr));
		
		
		int exitCode = shell.execute("echo 33");
		
		assertThat(exitCode, is(0));
		
		String outMessage = stdout.toString().trim();
		System.out.println(outMessage);
	}

	private OutputStream stream(StringWriter out) {
		return new WriterOutputStream(out);
	}
}
