package net.benelog.shellproxy.servlet;
import java.io.IOException;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;


public class StopServlet extends HttpServlet {

	private static final long serialVersionUID = 2229123705581054665L;

	protected void doGet(HttpServletRequest request,
			HttpServletResponse response) throws ServletException, IOException {
		System.exit(0);
		// This code should not be executed in test codes.
		// It may halts Maven builds or CI server.
	}
}
