#version 330

uniform sampler2DArray tex;
uniform float threshold;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, vec3(tex_coord, 0));
    float brightness = dot(in_color.rgb, vec3(0.2126, 0.7152, 0.0722));

    if (brightness > threshold) {
        color = in_color;
    } else {
        color = vec4(0.0);
    }
}