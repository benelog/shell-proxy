/*
 * @(#)PreviewServletTest.java $version 2012. 8. 31.
 *
 * Copyright 2007 NHN Corp. All rights Reserved. 
 * NHN PROPRIETARY/CONFIDENTIAL. Use is subject to license terms.
 */

package net.benelog.shellproxy.servlet;

import static org.junit.Assert.*;

import java.io.IOException;

import javax.servlet.ServletException;

import org.junit.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;


/**
 * @author benelog
 */
public class ExecutionServletTest {
	
	ExecutionServlet servlet = new ExecutionServlet(); 
	MockHttpServletRequest request = new MockHttpServletRequest();
	MockHttpServletResponse response = new MockHttpServletResponse();
	
	@Test
	public void getRequsetShouldBeAccepted() throws ServletException,IOException {
		request.addParameter("command", "echo hello");

		servlet.doGet(request, response);

		String contents = response.getContentAsString();
		System.out.println(contents);
		assertTrue(contents.contains("hello"));
	}
}
