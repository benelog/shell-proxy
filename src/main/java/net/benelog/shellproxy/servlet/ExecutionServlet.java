package net.benelog.shellproxy.servlet;
import java.io.IOException;
import java.io.OutputStream;
import java.io.StringWriter;
import java.io.Writer;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import net.benelog.shellproxy.executor.ProcessExecutionResult;
import net.benelog.shellproxy.executor.ShellExecutor;

import org.apache.commons.io.IOUtils;
import org.codehaus.jackson.map.ObjectMapper;
import org.eclipse.jetty.io.WriterOutputStream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;



public class ExecutionServlet extends HttpServlet {

	private static final long serialVersionUID = 2229123705581054665L;
	private static final int timeoutMillsecs = 61000;
	private final Logger log = LoggerFactory.getLogger(ExecutionServlet.class);
	private ObjectMapper mapper = new ObjectMapper();
	
	private OutputStream stream(Writer out) {
		return new WriterOutputStream(out);
	}
	
	protected void doGet(HttpServletRequest request,HttpServletResponse response) throws ServletException, IOException {
		String command = request.getParameter("command");
		
		ProcessExecutionResult execResult = execute(command);
		log.info("result: {}", execResult);
		
		String jsonResponse = mapper.writeValueAsString(execResult);
		response.getWriter().print(jsonResponse);
	}

	private ProcessExecutionResult execute(String command) {
		StringWriter processOut = new StringWriter();
		StringWriter processError = new StringWriter();

		OutputStream outStream = stream(processOut);
		OutputStream errorStream = stream(processError);
		
		ShellExecutor executor = new ShellExecutor(timeoutMillsecs, outStream, errorStream);
		int exitCode = executor.execute(command);

		
		ProcessExecutionResult execResult = new ProcessExecutionResult();
		execResult.setExitCode(exitCode);
		execResult.setStandardOutput(processOut.toString());
		execResult.setStandardError(processError.toString());
		
		IOUtils.closeQuietly(outStream); 
		IOUtils.closeQuietly(errorStream);
		return execResult;
	}
}
