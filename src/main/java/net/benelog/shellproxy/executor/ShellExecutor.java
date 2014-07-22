package net.benelog.shellproxy.executor;

import java.io.IOException;
import java.io.OutputStream;

import net.jcip.annotations.NotThreadSafe;

import org.apache.commons.exec.CommandLine;
import org.apache.commons.exec.DefaultExecutor;
import org.apache.commons.exec.ExecuteException;
import org.apache.commons.exec.ExecuteStreamHandler;
import org.apache.commons.exec.ExecuteWatchdog;
import org.apache.commons.exec.PumpStreamHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * @author benelog
 */
@NotThreadSafe
public class ShellExecutor {

	private static final Logger log = LoggerFactory.getLogger(ShellExecutor.class);
	private int timeoutMillsecs = 0;
	private OutputStream out;
	private OutputStream err;
	
	public ShellExecutor(){
	}

	/**
	 * @param timeoutMillsecs
	 */
	public ShellExecutor(int timeoutMillsecs) {
		this.timeoutMillsecs = timeoutMillsecs;
	}

	/**
	 * @param timeoutMillsecs
	 */
	public ShellExecutor(int timeoutMillsecs, OutputStream out, OutputStream err) {
		this.timeoutMillsecs = timeoutMillsecs;
		this.out = out;
		this.err = err;
	}

	/**
	 * @param options
	 * @return
	 */
	public int execute(String command) {
		DefaultExecutor executor = new DefaultExecutor();
		CommandLine cmdLine = CommandLine.parse(command);
		log.info("execute command : {}", cmdLine);
		
		if (timeoutMillsecs > 0 ) {
			ExecuteWatchdog watchdog = new ExecuteWatchdog(this.timeoutMillsecs);
			executor.setWatchdog(watchdog);
		}
		if (out != null) {
			ExecuteStreamHandler streamHandler = new PumpStreamHandler(out, err);
			executor.setStreamHandler(streamHandler);
		}
		try {
			return executor.execute(cmdLine);
		} catch (ExecuteException e) {
			e.printStackTrace();
			return e.getExitValue();
		} catch (IOException e) {
			throw new IllegalStateException(e);
		}	
	}
}
